import type { Theme } from '../types';
import { SHOKA_LIGHT, SHOKA_DARK } from './shoka';
import { ROSE_PINE, ROSE_PINE_MOON, ROSE_PINE_DAWN } from './rose-pine';
import {
	CATPPUCCIN_LATTE,
	CATPPUCCIN_FRAPPE,
	CATPPUCCIN_MACCHIATO,
	CATPPUCCIN_MOCHA
} from './catppuccin';
import { KANAGAWA_WAVE, KANAGAWA_DRAGON, KANAGAWA_LOTUS } from './kanagawa';
import { GRUVBOX_DARK, GRUVBOX_LIGHT } from './gruvbox';
import { NORD } from './nord';
import { TOKYO_NIGHT } from './tokyo-night';

export const DEFAULT_THEME_ID = 'shoka-dark';
export const DEFAULT_LIGHT_THEME_ID = 'shoka-light';

/** Every built-in theme, in menu order. */
export const PRESETS: Theme[] = [
	SHOKA_DARK,
	SHOKA_LIGHT,
	ROSE_PINE,
	ROSE_PINE_MOON,
	ROSE_PINE_DAWN,
	CATPPUCCIN_LATTE,
	CATPPUCCIN_FRAPPE,
	CATPPUCCIN_MACCHIATO,
	CATPPUCCIN_MOCHA,
	KANAGAWA_WAVE,
	KANAGAWA_DRAGON,
	KANAGAWA_LOTUS,
	GRUVBOX_DARK,
	GRUVBOX_LIGHT,
	NORD,
	TOKYO_NIGHT
];

export const PRESET_MAP: Record<string, Theme> = Object.fromEntries(
	PRESETS.map((theme) => [theme.id, theme])
);

export {
	SHOKA_LIGHT,
	SHOKA_DARK,
	ROSE_PINE,
	ROSE_PINE_MOON,
	ROSE_PINE_DAWN,
	CATPPUCCIN_LATTE,
	CATPPUCCIN_FRAPPE,
	CATPPUCCIN_MACCHIATO,
	CATPPUCCIN_MOCHA,
	KANAGAWA_WAVE,
	KANAGAWA_DRAGON,
	KANAGAWA_LOTUS,
	GRUVBOX_DARK,
	GRUVBOX_LIGHT,
	NORD,
	TOKYO_NIGHT
};
