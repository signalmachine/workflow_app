export async function fileToBase64(file: File): Promise<string> {
	return new Promise((resolve, reject) => {
		const reader = new FileReader();
		reader.onerror = () => reject(reader.error ?? new Error('Failed to read file'));
		reader.onload = () => {
			const result = reader.result;
			if (typeof result !== 'string') {
				reject(new Error('Failed to read file'));
				return;
			}
			const commaIndex = result.indexOf(',');
			resolve(commaIndex >= 0 ? result.slice(commaIndex + 1) : result);
		};
		reader.readAsDataURL(file);
	});
}
