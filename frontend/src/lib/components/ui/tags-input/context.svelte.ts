import { getContext, setContext } from 'svelte';

type Getter<T> = () => T;

export type TagsInputRootStateProps = {
	value: Getter<string[]>;
	onValueChange: (value: string[]) => void;
	max: Getter<number | undefined>;
	duplicate: Getter<boolean>;
	delimiter: Getter<string | RegExp>;
	addOnPaste: Getter<boolean>;
	addOnTab: Getter<boolean>;
	disabled: Getter<boolean>;
};

export class TagsInputRootState {
	#props: TagsInputRootStateProps;

	activeIndex = $state<number | null>(null);
	inputRef = $state<HTMLInputElement | null>(null);
	#elements = new Map<string, HTMLElement>();

	constructor(props: TagsInputRootStateProps) {
		this.#props = props;
	}

	get value() {
		return this.#props.value();
	}
	get disabled() {
		return this.#props.disabled();
	}
	get duplicate() {
		return this.#props.duplicate();
	}
	get delimiter() {
		return this.#props.delimiter();
	}
	get max() {
		return this.#props.max();
	}
	get addOnPaste() {
		return this.#props.addOnPaste();
	}
	get addOnTab() {
		return this.#props.addOnTab();
	}
	get isMaxReached() {
		const max = this.max;
		return max !== undefined && this.value.length >= max;
	}

	/** Attempts to add a single tag, returns whether it was added. */
	addTag(raw: string): boolean {
		if (this.disabled) return false;
		const text = raw.trim();
		if (!text) return false;
		if (this.isMaxReached) return false;
		if (!this.duplicate && this.value.includes(text)) return false;
		this.#props.onValueChange([...this.value, text]);
		return true;
	}

	/** Splits raw text on the configured delimiter and adds each part as a tag. */
	addFromDelimited(raw: string): boolean {
		const parts = raw.split(this.delimiter);
		let added = false;
		for (const part of parts) {
			if (this.addTag(part)) added = true;
		}
		return added;
	}

	removeTagAt(index: number) {
		if (this.disabled) return;
		const next = this.value.slice();
		if (index < 0 || index >= next.length) return;
		next.splice(index, 1);
		this.#props.onValueChange(next);
	}

	clear() {
		if (this.disabled) return;
		this.#props.onValueChange([]);
	}

	registerElement(value: string, el: HTMLElement) {
		this.#elements.set(value, el);
		return () => {
			if (this.#elements.get(value) === el) this.#elements.delete(value);
		};
	}

	focusInput() {
		this.activeIndex = null;
		this.inputRef?.focus();
	}

	focusItemAt(index: number) {
		const values = this.value;
		if (index < 0 || index >= values.length) {
			this.focusInput();
			return;
		}
		const el = this.#elements.get(values[index]);
		this.activeIndex = index;
		el?.focus();
	}

	focusLast() {
		this.focusItemAt(this.value.length - 1);
	}
}

export class TagsInputItemState {
	root: TagsInputRootState;
	#itemValue: Getter<string>;
	#itemDisabled: Getter<boolean>;

	constructor(root: TagsInputRootState, itemValue: Getter<string>, itemDisabled: Getter<boolean>) {
		this.root = root;
		this.#itemValue = itemValue;
		this.#itemDisabled = itemDisabled;
	}

	get value() {
		return this.#itemValue();
	}
	get index() {
		return this.root.value.indexOf(this.value);
	}
	get disabled() {
		return this.root.disabled || this.#itemDisabled();
	}
	get active() {
		return this.index !== -1 && this.root.activeIndex === this.index;
	}
}

const ROOT_KEY = Symbol('tags-input-root');
const ITEM_KEY = Symbol('tags-input-item');

export function setTagsInputRoot(props: TagsInputRootStateProps) {
	return setContext(ROOT_KEY, new TagsInputRootState(props));
}

export function useTagsInputRoot(): TagsInputRootState {
	return getContext(ROOT_KEY);
}

export function setTagsInputItem(
	root: TagsInputRootState,
	itemValue: Getter<string>,
	itemDisabled: Getter<boolean>
) {
	return setContext(ITEM_KEY, new TagsInputItemState(root, itemValue, itemDisabled));
}

export function useTagsInputItem(): TagsInputItemState {
	return getContext(ITEM_KEY);
}
