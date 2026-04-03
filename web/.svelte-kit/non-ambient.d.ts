
// this file is generated — do not edit it


declare module "svelte/elements" {
	export interface HTMLAttributes<T> {
		'data-sveltekit-keepfocus'?: true | '' | 'off' | undefined | null;
		'data-sveltekit-noscroll'?: true | '' | 'off' | undefined | null;
		'data-sveltekit-preload-code'?:
			| true
			| ''
			| 'eager'
			| 'viewport'
			| 'hover'
			| 'tap'
			| 'off'
			| undefined
			| null;
		'data-sveltekit-preload-data'?: true | '' | 'hover' | 'tap' | 'off' | undefined | null;
		'data-sveltekit-reload'?: true | '' | 'off' | undefined | null;
		'data-sveltekit-replacestate'?: true | '' | 'off' | undefined | null;
	}
}

export {};


declare module "$app/types" {
	type MatcherParam<M> = M extends (param : string) => param is (infer U extends string) ? U : string;

	export interface AppTypes {
		RouteId(): "/(public)" | "/(app)" | "/" | "/(app)/admin" | "/(app)/admin/access" | "/(app)/admin/accounting" | "/(app)/admin/inventory" | "/(app)/admin/parties" | "/(app)/inventory" | "/(public)/login" | "/(app)/operations" | "/(app)/review" | "/(app)/settings";
		RouteParams(): {
			
		};
		LayoutParams(): {
			"/(public)": Record<string, never>;
			"/(app)": Record<string, never>;
			"/": Record<string, never>;
			"/(app)/admin": Record<string, never>;
			"/(app)/admin/access": Record<string, never>;
			"/(app)/admin/accounting": Record<string, never>;
			"/(app)/admin/inventory": Record<string, never>;
			"/(app)/admin/parties": Record<string, never>;
			"/(app)/inventory": Record<string, never>;
			"/(public)/login": Record<string, never>;
			"/(app)/operations": Record<string, never>;
			"/(app)/review": Record<string, never>;
			"/(app)/settings": Record<string, never>
		};
		Pathname(): "/" | "/admin" | "/admin/access" | "/admin/accounting" | "/admin/inventory" | "/admin/parties" | "/inventory" | "/login" | "/operations" | "/review" | "/settings";
		ResolvedPathname(): `${"" | `/${string}`}${ReturnType<AppTypes['Pathname']>}`;
		Asset(): "/favicon.svg" | string & {};
	}
}