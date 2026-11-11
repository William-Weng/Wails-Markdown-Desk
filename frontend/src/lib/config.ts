import { GetConfig } from "../../../frontend/bindings/MarkdownDesk/backend/configservice";

export interface AppConfig {
	imageWebRoute: string;
}

export let appConfig: AppConfig | null = null;

export async function loadConfig(): Promise<AppConfig> {
	const config = await GetConfig();
	appConfig = config;

	return config;
}

export async function getConfig(): Promise<AppConfig> {
	if (!appConfig) {
		await loadConfig();
	}

	return appConfig!;
}
