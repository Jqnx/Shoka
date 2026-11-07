// Single playback controller shared by every route. The audio element is
// created once with `new Audio()` and never inserted into the DOM, so
// playback survives client-side navigation; the UI only mirrors its state.
import { browser } from '$app/environment';
import { audioTrackSrc } from '$lib/audio';
import { audioCoverSrc, type Audio, type AudioTrack } from '$lib/types';

export type RepeatMode = 'off' | 'all' | 'one';

export type PlayerSource = {
	audioId: string;
	updatedAt: string;
	title: string;
	circle: string | null;
	coverUrl: string | null;
	folder: string;
	tracks: AudioTrack[];
};

export const PLAYBACK_RATES = [0.75, 1, 1.25, 1.5, 2];

export function sourceFromAudio(audio: Audio, folder: string): PlayerSource {
	return {
		audioId: audio.id,
		updatedAt: audio.updated_at,
		title: audio.title,
		circle: audio.circle,
		coverUrl: audio.cover_ready ? audioCoverSrc(audio) : null,
		folder,
		tracks: audio.tracks.filter((track) => track.folder === folder)
	};
}

const SETTINGS_KEY = 'shoka.player.settings';
const POSITIONS_KEY = 'shoka.player.positions';

type StoredSettings = {
	volume: number;
	muted: boolean;
};

const SAVE_INTERVAL_MS = 5000;

class PlayerStore {
	source = $state<PlayerSource | null>(null);
	trackIndex = $state<number | null>(null);
	playing = $state(false);
	loading = $state(false);
	error = $state<string | null>(null);
	currentTime = $state(0);
	duration = $state(0);
	volume = $state(1);
	muted = $state(false);
	rate = $state(1);
	repeat = $state<RepeatMode>('off');
	shuffle = $state(false);
	// Expanded player card is open; the mini bar is hidden while it is.
	expanded = $state(false);
	// Stop dismisses the player UI while keeping the loaded track and queue.
	barVisible = $state(true);

	#el: HTMLAudioElement | null = null;
	#generation = 0;
	#positions: Record<string, number> = {};
	#lastSavedAt = 0;

	#onPageHide = () => {
		this.#savePosition(true);
	};

	get track(): AudioTrack | null {
		const source = this.source;
		if (!source || this.trackIndex === null) return null;
		return source.tracks.find((track) => track.index === this.trackIndex) ?? null;
	}

	get position(): number {
		const source = this.source;
		if (!source || this.trackIndex === null) return -1;
		return source.tracks.findIndex((track) => track.index === this.trackIndex);
	}

	playSource(source: PlayerSource, trackIndex?: number): void {
		this.#ensureElement();
		if (!this.#el) return;
		if (source.tracks.length === 0) return;

		const index = trackIndex ?? source.tracks[0].index;
		// Force-save before switching sources: the old time must stay under
		// the old work's key, not be written under the new one.
		this.#savePosition(true);
		this.source = source;
		this.barVisible = true;
		this.#load(index, true);
	}

	toggle(): void {
		const el = this.#el;
		if (!el) return;
		if (el.paused) this.play();
		else this.pause();
	}

	play(): void {
		const el = this.#el;
		if (!el) return;
		this.barVisible = true;
		void el.play().catch((error: DOMException) => {
			if (error.name === 'AbortError') return;
			this.error = 'Unable to play this track.';
			this.playing = false;
		});
	}

	pause(): void {
		this.#el?.pause();
	}

	stop(): void {
		const el = this.#el;
		if (!el) return;

		// Invalidate a pending loadedmetadata so stop cannot be undone by an
		// in-flight track switch, then reset to the start and stay paused.
		this.#generation++;
		el.pause();
		try {
			el.currentTime = 0;
		} catch {
			// Some browsers reject writes before metadata is available.
		}
		this.currentTime = 0;
		this.playing = false;
		this.loading = false;
		this.error = null;
		this.#savePosition(true);
		this.#setPlaybackState();
		this.#updatePositionState();
		this.expanded = false;
		this.barVisible = false;
	}

	seek(seconds: number): void {
		const el = this.#el;
		if (!el || !Number.isFinite(seconds)) return;
		if (!Number.isFinite(el.duration) || el.duration <= 0) return;

		const target = Math.min(Math.max(seconds, 0), el.duration);
		el.currentTime = target;
		this.currentTime = target;
		this.#savePosition(true);
		this.#updatePositionState();
	}

	next(): void {
		const next = this.#nextIndex();
		if (next === null) return;

		this.#savePosition(true);
		this.#load(next, false);
	}

	previous(): void {
		if (!this.#el) return;
		if (this.currentTime > 3) {
			this.seek(0);
			return;
		}

		const source = this.source;
		if (!source) return;

		const position = this.position;
		let target: number | null = null;

		if (this.shuffle && source.tracks.length > 1) {
			const others = source.tracks.filter((_, index) => index !== position);
			target = others[Math.floor(Math.random() * others.length)].index;
		} else if (position > 0) {
			target = source.tracks[position - 1].index;
		} else if (this.repeat === 'all') {
			target = source.tracks[source.tracks.length - 1].index;
		}

		if (target === null) {
			this.seek(0);
			return;
		}

		this.#savePosition(true);
		this.#load(target, false);
	}

	setVolume(value: number): void {
		if (!Number.isFinite(value)) return;

		const volume = Math.min(Math.max(value, 0), 1);
		this.volume = volume;
		const el = this.#el;
		if (el) {
			el.volume = volume;
			if (volume > 0) el.muted = false;
		}
		if (volume > 0) this.muted = false;
		this.#persistSettings();
	}

	toggleMute(): void {
		this.muted = !this.muted;
		if (this.#el) this.#el.muted = this.muted;
		this.#persistSettings();
	}

	setRate(value: number): void {
		if (!Number.isFinite(value) || !PLAYBACK_RATES.includes(value)) return;

		this.rate = value;
		const el = this.#el;
		if (el) {
			// load() resets playbackRate to defaultPlaybackRate, so both must
			// change or the next track silently falls back to 1×.
			el.defaultPlaybackRate = value;
			el.playbackRate = value;
		}
		this.#persistSettings();
	}

	cycleRepeat(): void {
		this.repeat = this.repeat === 'off' ? 'all' : this.repeat === 'all' ? 'one' : 'off';
		this.#persistSettings();
	}

	toggleShuffle(): void {
		this.shuffle = !this.shuffle;
		this.#persistSettings();
	}

	isCurrent(audioId: string, updatedAt: string, folder: string, trackIndex?: number): boolean {
		const source = this.source;
		if (!source) return false;
		if (source.audioId !== audioId || source.updatedAt !== updatedAt) return false;
		if (source.folder !== folder) return false;
		if (trackIndex !== undefined && trackIndex !== this.trackIndex) return false;
		return true;
	}

	destroy(): void {
		this.#savePosition(true);

		const el = this.#el;
		if (el) {
			el.src = '';
			el.load();
		}

		this.#generation++;
		if (browser) window.removeEventListener('pagehide', this.#onPageHide);
		this.#el = null;
	}

	#ensureElement(): void {
		if (!browser || this.#el) return;

		const el = new Audio();
		el.preload = 'metadata';

		el.addEventListener('timeupdate', () => {
			this.currentTime = el.currentTime;
			if (Number.isFinite(el.duration) && el.duration > 0) this.duration = el.duration;
			this.#savePosition(false);
			this.#updatePositionState();
		});

		el.addEventListener('durationchange', () => {
			if (Number.isFinite(el.duration) && el.duration > 0) this.duration = el.duration;
		});

		el.addEventListener('play', () => {
			this.playing = true;
			this.#setPlaybackState();
		});

		el.addEventListener('pause', () => {
			this.playing = false;
			this.#savePosition(true);
			this.#setPlaybackState();
		});

		el.addEventListener('ended', () => {
			this.playing = false;
			const source = this.source;
			const track = this.track;
			if (source && track) {
				delete this.#positions[this.#positionKey(source, track)];
				this.#persistPositions();
			}
			this.#handleEnded();
		});

		el.addEventListener('waiting', () => {
			this.loading = true;
		});

		el.addEventListener('playing', () => {
			this.playing = true;
			this.loading = false;
		});

		el.addEventListener('canplay', () => {
			this.loading = false;
		});

		el.addEventListener('volumechange', () => {
			this.volume = el.volume;
			this.muted = el.muted;
			this.#persistSettings();
		});

		el.addEventListener('ratechange', () => {
			this.rate = el.playbackRate;
			this.#persistSettings();
		});

		el.addEventListener('error', () => {
			this.error = 'Unable to play this track.';
			this.playing = false;
			this.loading = false;
		});

		this.#el = el;
		this.#loadStored();
		el.volume = this.volume;
		el.muted = this.muted;
		el.defaultPlaybackRate = this.rate;
		el.playbackRate = this.rate;
		this.#setupMediaSession();
		window.addEventListener('pagehide', this.#onPageHide);
	}

	#load(trackIndex: number, resume: boolean): void {
		const source = this.source;
		const el = this.#el;
		if (!source || !el) return;

		const track = source.tracks.find((candidate) => candidate.index === trackIndex);
		if (!track) return;

		this.trackIndex = trackIndex;
		this.currentTime = 0;
		this.error = null;
		this.loading = true;
		// Any path that starts a track (media next/previous included) must
		// restore the player UI; stop() hides it but keeps the source.
		this.barVisible = true;
		this.duration =
			Number.isFinite(track.duration_seconds) && track.duration_seconds > 0
				? track.duration_seconds
				: 0;

		const generation = ++this.#generation;
		const saved = resume ? (this.#positions[this.#positionKey(source, track)] ?? 0) : 0;

		// Attach before assigning src: a cached file can fire loadedmetadata
		// in the same task as load().
		const onLoadedMetadata = () => {
			el.removeEventListener('loadedmetadata', onLoadedMetadata);
			if (generation !== this.#generation) return;

			if (Number.isFinite(el.duration) && el.duration > 0) this.duration = el.duration;
			if (saved > 5 && saved < el.duration - 10) {
				el.currentTime = saved;
				this.currentTime = saved;
			}
			this.loading = false;
			this.#setMetadata(track, source);
		};
		el.addEventListener('loadedmetadata', onLoadedMetadata);

		el.src = audioTrackSrc(source.audioId, trackIndex, source.updatedAt);
		el.load();

		// Keep the click's user activation; metadata normally arrives first.
		void el.play().catch((error: DOMException) => {
			if (error.name === 'AbortError') return;
			this.error = 'Unable to play this track.';
			this.loading = false;
		});
	}

	#nextIndex(): number | null {
		const source = this.source;
		if (!source) return null;

		const position = this.position;
		if (position === -1) return null;

		const tracks = source.tracks;
		if (tracks.length === 0) return null;
		if (tracks.length === 1) {
			return this.repeat === 'all' ? tracks[0].index : null;
		}

		if (this.shuffle) {
			const others = tracks.filter((_, index) => index !== position);
			return others[Math.floor(Math.random() * others.length)].index;
		}

		if (position < tracks.length - 1) return tracks[position + 1].index;
		if (this.repeat === 'all') return tracks[0].index;
		return null;
	}

	#handleEnded(): void {
		if (this.repeat === 'one') {
			this.seek(0);
			this.play();
			return;
		}

		const next = this.#nextIndex();
		if (next !== null) {
			this.#load(next, false);
			return;
		}

		this.pause();
		const el = this.#el;
		if (el) el.currentTime = 0;
		this.currentTime = 0;
	}

	#positionKey(source: PlayerSource, track: AudioTrack): string {
		return `${source.audioId}:${source.folder}:${track.name}`;
	}

	#savePosition(force: boolean): void {
		if (!browser) return;

		const el = this.#el;
		const source = this.source;
		if (!el || !source || this.trackIndex === null) return;
		// A new src fires pause/timeupdate before it is ready; saving then
		// would overwrite the outgoing track with 0. `ended` guard stops the
		// final pause from re-adding the position `ended` just deleted.
		if (this.loading || el.ended) return;

		const now = Date.now();
		if (!force && now - this.#lastSavedAt < SAVE_INTERVAL_MS) return;

		const track = this.track;
		if (!track) return;

		this.#positions[this.#positionKey(source, track)] = el.currentTime;
		this.#lastSavedAt = now;
		this.#persistPositions();
	}

	#loadStored(): void {
		if (!browser) return;

		try {
			const raw = localStorage.getItem(SETTINGS_KEY);
			if (raw) {
				const stored = JSON.parse(raw) as Partial<StoredSettings>;
				if (typeof stored.volume === 'number' && Number.isFinite(stored.volume)) {
					this.volume = Math.min(Math.max(stored.volume, 0), 1);
				}
				if (typeof stored.muted === 'boolean') this.muted = stored.muted;
				// rate/repeat/shuffle have no UI right now; ignoring stale
				// values keeps a previously stored 1.5×/repeat-one from
				// becoming an unreachable mode.
			}
		} catch {
			// Storage can be unavailable (private mode/quota); defaults are fine.
		}

		// Rewrite the (possibly legacy) blob so hidden rate/repeat/shuffle
		// keys do not linger in storage.
		this.#persistSettings();

		try {
			const raw = localStorage.getItem(POSITIONS_KEY);
			if (raw) {
				const stored = JSON.parse(raw) as Record<string, unknown>;
				if (stored && typeof stored === 'object') {
					const positions: Record<string, number> = {};
					for (const [key, value] of Object.entries(stored)) {
						if (typeof value === 'number' && Number.isFinite(value)) positions[key] = value;
					}
					this.#positions = positions;
				}
			}
		} catch {
			// Ignore malformed or unavailable stored positions.
		}
	}

	#persistSettings(): void {
		if (!browser) return;

		const settings: StoredSettings = {
			volume: this.volume,
			muted: this.muted
		};

		try {
			localStorage.setItem(SETTINGS_KEY, JSON.stringify(settings));
		} catch {
			// Persistence is best-effort.
		}
	}

	#persistPositions(): void {
		if (!browser) return;

		try {
			localStorage.setItem(POSITIONS_KEY, JSON.stringify(this.#positions));
		} catch {
			// Persistence is best-effort.
		}
	}

	#setupMediaSession(): void {
		if (!browser || !('mediaSession' in navigator)) return;

		const setHandler = (action: MediaSessionAction, handler: MediaSessionActionHandler) => {
			try {
				navigator.mediaSession.setActionHandler(action, handler);
			} catch {
				// Browsers reject unsupported actions.
			}
		};

		setHandler('play', () => this.play());
		setHandler('pause', () => this.pause());
		setHandler('previoustrack', () => this.previous());
		setHandler('nexttrack', () => this.next());
		setHandler('seekbackward', (details) => {
			this.seek(this.currentTime - (details.seekOffset ?? 10));
		});
		setHandler('seekforward', (details) => {
			this.seek(this.currentTime + (details.seekOffset ?? 10));
		});
		setHandler('seekto', (details) => {
			if (typeof details.seekTime === 'number') this.seek(details.seekTime);
		});
	}

	#setMetadata(track: AudioTrack, source: PlayerSource): void {
		if (!browser || !('mediaSession' in navigator)) return;

		try {
			navigator.mediaSession.metadata = new MediaMetadata({
				title: track.name,
				artist: source.title,
				album: source.circle ?? 'Shoka',
				artwork: source.coverUrl ? [{ src: new URL(source.coverUrl, location.origin).href }] : []
			});
		} catch {
			// Media Session metadata is optional.
		}
		this.#setPlaybackState();
	}

	#setPlaybackState(): void {
		if (!browser || !('mediaSession' in navigator)) return;

		try {
			navigator.mediaSession.playbackState = this.playing ? 'playing' : 'paused';
		} catch {
			// Media Session is optional.
		}
	}

	#updatePositionState(): void {
		if (!browser || !('mediaSession' in navigator)) return;
		if (!Number.isFinite(this.duration) || this.duration <= 0) return;

		try {
			navigator.mediaSession.setPositionState({
				duration: this.duration,
				playbackRate: this.rate,
				position: Math.min(this.currentTime, this.duration)
			});
		} catch {
			// Position state is advisory.
		}
	}
}

export const player = new PlayerStore();
if (import.meta.hot) import.meta.hot.dispose(() => player.destroy());
