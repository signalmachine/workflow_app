<script lang="ts">
	interface Props {
		userDisplayName: string;
		orgName: string;
		roleCode: string;
		navExpanded: boolean;
		onToggleNav: () => void;
		onLogout: () => void;
	}

	let { userDisplayName, orgName, roleCode, navExpanded, onToggleNav, onLogout }: Props = $props();
</script>

<header class="topbar">
	<div class="brand-row">
		<button
			aria-label={navExpanded ? 'Collapse navigation' : 'Expand navigation'}
			aria-pressed={navExpanded}
			class="menu-button"
			onclick={onToggleNav}
			type="button"
		>
			<span></span>
			<span></span>
			<span></span>
		</button>
		<div>
			<div class="brand-mark">Workflow App</div>
			<div class="org-name">{orgName}</div>
		</div>
	</div>
	<div class="user-row">
		<div class="user-copy">
			<div>{userDisplayName}</div>
			<div class="role-hint">{roleCode.replaceAll('_', ' ')}</div>
		</div>
		<button class="logout-button" onclick={onLogout} type="button">Sign out</button>
	</div>
</header>

<style>
	.topbar {
		align-items: center;
		background: var(--shell-ink-strong);
		color: #f6f8fa;
		display: flex;
		height: 48px;
		justify-content: space-between;
		left: 0;
		padding: 0 1.15rem;
		position: sticky;
		top: 0;
		z-index: 20;
		box-shadow: 0 10px 24px rgba(7, 18, 26, 0.14);
	}

	.brand-row,
	.user-row {
		align-items: center;
		display: flex;
		gap: 0.875rem;
	}

	.brand-mark {
		font-size: var(--text-sm);
		font-weight: 600;
		letter-spacing: 0.04em;
		text-transform: uppercase;
	}

	.org-name,
	.role-hint {
		color: rgba(255, 255, 255, 0.68);
		font-size: var(--text-2xs);
	}

	.menu-button,
	.logout-button {
		background: transparent;
		border: 1px solid rgba(255, 255, 255, 0.14);
		border-radius: 999px;
		color: inherit;
	}

	.menu-button {
		padding: 0.45rem 0.55rem;
	}

	.menu-button span {
		background: currentColor;
		display: block;
		height: 2px;
		margin: 3px 0;
		width: 14px;
	}

	.logout-button {
		background: rgba(255, 255, 255, 0.06);
		padding: 0.45rem 0.82rem;
	}

	.user-copy {
		text-align: right;
	}

	@media (max-width: 767px) {
		.user-copy {
			display: none;
		}
	}
</style>
