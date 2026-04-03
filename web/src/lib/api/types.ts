export interface APIError {
	error: string;
}

export interface SessionContext {
	session_id: string;
	org_id: string;
	org_slug: string;
	org_name: string;
	user_id: string;
	user_email: string;
	user_display_name: string;
	role_code: string;
	device_label: string;
	expires_at: string;
	issued_at: string;
	last_seen_at: string;
}

export interface SessionLoginRequest {
	org_slug: string;
	email: string;
	password: string;
	device_label: string;
}
