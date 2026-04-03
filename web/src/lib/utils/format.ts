export function formatDateTime(value?: string): string {
	if (!value) {
		return '-';
	}
	return new Date(value).toLocaleString(undefined, {
		year: 'numeric',
		month: 'short',
		day: '2-digit',
		hour: '2-digit',
		minute: '2-digit'
	});
}

export function formatDate(value?: string): string {
	if (!value) {
		return '-';
	}
	return new Date(value).toLocaleDateString(undefined, {
		year: 'numeric',
		month: 'short',
		day: '2-digit'
	});
}

export function formatMinorUnits(value?: number): string {
	if (value === undefined) {
		return '-';
	}
	return new Intl.NumberFormat(undefined, {
		style: 'currency',
		currency: 'INR',
		maximumFractionDigits: 2,
		minimumFractionDigits: 2
	}).format(value / 100);
}

export function formatMilliQuantity(value?: number): string {
	if (value === undefined) {
		return '-';
	}
	return `${(value / 1000).toLocaleString(undefined, { maximumFractionDigits: 3 })}`;
}

export function humanizeStatus(value: string): string {
	return value
		.replaceAll('_', ' ')
		.replace(/\b\w/g, (match) => match.toUpperCase());
}
