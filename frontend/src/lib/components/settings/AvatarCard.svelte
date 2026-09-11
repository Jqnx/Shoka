<script lang="ts">
	import * as Card from '$lib/components/ui/card/index.js';
	import * as Avatar from '$lib/components/ui/avatar/index.js';
	import { Button } from '$lib/components/ui/button/index.js';
	import AvatarCropDialog from './AvatarCropDialog.svelte';
	import { enhance } from '$app/forms';
	import { invalidateAll } from '$app/navigation';
	import type { SubmitFunction } from '@sveltejs/kit';
	import { cn } from '$lib/utils.js';
	import { ACCEPTED_MIME_TYPES, MAX_SOURCE_BYTES } from '$lib/schemas/avatar';
	import type { User } from 'better-auth';

	let { user }: { user: User } = $props();

	const initials = $derived(
		user.name
			.split(' ')
			.map((part: string) => part[0])
			.slice(0, 2)
			.join('')
			.toUpperCase()
	);

	let fileInput = $state<HTMLInputElement>();
	let formEl = $state<HTMLFormElement>();

	let cropOpen = $state(false);
	let cropImage = $state<string | null>(null);
	// Bumped per file pick so `{#key}` remounts the dialog with fresh crop/zoom
	// state instead of carrying over the previous image's values.
	let cropNonce = $state(0);

	let dragOver = $state(false);
	let uploading = $state(false);
	let removing = $state(false);
	let avatarError = $state<string | null>(null);

	function openPicker() {
		fileInput?.click();
	}

	function acceptFile(file: File | undefined) {
		if (!file) return;
		if (!ACCEPTED_MIME_TYPES.includes(file.type)) {
			avatarError = 'Use JPEG, PNG, WebP, or AVIF.';
			return;
		}
		if (file.size > MAX_SOURCE_BYTES) {
			avatarError = 'Image is too large.';
			return;
		}

		avatarError = null;
		if (cropImage) URL.revokeObjectURL(cropImage);
		cropImage = URL.createObjectURL(file);
		cropNonce += 1;
		cropOpen = true;
	}

	function handleInputChange(e: Event) {
		const target = e.currentTarget as HTMLInputElement;
		acceptFile(target.files?.[0]);
		// Reset so picking the same file again still fires a change event.
		target.value = '';
	}

	function handleDrop(e: DragEvent) {
		e.preventDefault();
		dragOver = false;
		acceptFile(e.dataTransfer?.files?.[0]);
	}

	function handleCropped(blob: Blob) {
		if (!fileInput || !formEl) return;
		const dt = new DataTransfer();
		dt.items.add(new File([blob], 'avatar.webp', { type: 'image/webp' }));
		fileInput.files = dt.files;
		formEl.requestSubmit();
	}

	// The object URL for whichever file is currently loaded into the crop
	// dialog - revoke it once, on destroy, rather than after every crop.
	$effect(() => {
		return () => {
			if (cropImage) URL.revokeObjectURL(cropImage);
		};
	});

	// A response that never reaches our action at all - e.g. adapter-node's
	// BODY_SIZE_LIMIT rejecting an oversized body with a bare 413 - lands here
	// as `result.type === 'error'` rather than 'failure', since there's no
	// ActionResult JSON to parse. Handle it explicitly rather than leaving the
	// user with no feedback at all.
	const handleUpload: SubmitFunction = () => {
		uploading = true;
		avatarError = null;
		return async ({ result }) => {
			uploading = false;
			if (result.type === 'failure') {
				avatarError =
					(result.data?.avatarError as string | undefined) ?? 'Failed to update avatar.';
				return;
			}
			if (result.type === 'error') {
				avatarError = 'Failed to update avatar.';
				return;
			}
			if (result.type === 'success') await invalidateAll();
		};
	};

	const handleRemove: SubmitFunction = () => {
		removing = true;
		avatarError = null;
		return async ({ result }) => {
			removing = false;
			if (result.type === 'failure') {
				avatarError =
					(result.data?.avatarError as string | undefined) ?? 'Failed to remove avatar.';
				return;
			}
			if (result.type === 'error') {
				avatarError = 'Failed to remove avatar.';
				return;
			}
			if (result.type === 'success') await invalidateAll();
		};
	};
</script>

<Card.Root class="mt-6">
	<Card.Header>
		<Card.Title>Avatar</Card.Title>
		<Card.Description>Shown in the sidebar and anywhere else your profile appears.</Card.Description
		>
	</Card.Header>
	<Card.Content>
		<div
			class="flex items-center gap-4"
			role="group"
			aria-label="Avatar upload"
			ondragover={(e) => {
				e.preventDefault();
				dragOver = true;
			}}
			ondragleave={() => (dragOver = false)}
			ondrop={handleDrop}
		>
			<Avatar.Root class="size-16">
				<Avatar.Image src={user.image ?? ''} alt={user.name} />
				<Avatar.Fallback class="text-lg">{initials}</Avatar.Fallback>
			</Avatar.Root>

			<div
				class={cn(
					'flex flex-1 flex-col gap-2 rounded-md border border-dashed p-3 text-sm text-muted-foreground',
					dragOver && 'border-primary bg-accent/50'
				)}
			>
				<p>Drag an image here, or</p>
				<div class="flex flex-wrap gap-2">
					<Button
						type="button"
						variant="outline"
						size="sm"
						onclick={openPicker}
						disabled={uploading}
					>
						{uploading ? 'Uploading…' : 'Choose image'}
					</Button>
					{#if user.image}
						<form method="POST" action="?/removeAvatar" use:enhance={handleRemove}>
							<Button type="submit" variant="ghost" size="sm" disabled={removing}>
								{removing ? 'Removing…' : 'Remove'}
							</Button>
						</form>
					{/if}
				</div>
			</div>
		</div>

		{#if avatarError}
			<p class="mt-2 text-sm text-destructive">{avatarError}</p>
		{/if}
	</Card.Content>
</Card.Root>

<form
	bind:this={formEl}
	method="POST"
	action="?/uploadAvatar"
	enctype="multipart/form-data"
	use:enhance={handleUpload}
	class="hidden"
>
	<input
		bind:this={fileInput}
		type="file"
		name="avatar"
		accept={ACCEPTED_MIME_TYPES.join(',')}
		onchange={handleInputChange}
	/>
</form>

{#if cropImage}
	{#key cropNonce}
		<AvatarCropDialog bind:open={cropOpen} image={cropImage} oncropped={handleCropped} />
	{/key}
{/if}
