import { RemoteDriver } from 'vuefinder';

/**
 * Custom RemoteDriver that transforms requests to use the Timeship v2 API format.
 * 
 * Converts paths from the VueFinder format (adapter://path) to the v2 API format:
 * /storages/{adapter}/nodes/{path}
 */
export class TimeshipRemoteDriver extends RemoteDriver {
  private adapter: string;
  private baseURL: string;
  private snapshot: string | null;

  constructor(config: any) {
    super(config);
    this.adapter = config.adapter || 'local';
    this.baseURL = config.baseURL || '';
    this.snapshot = config.snapshot || null;
    console.log('TimeshipRemoteDriver initialized with adapter:', this.adapter, 'baseURL:', this.baseURL, 'snapshot:', this.snapshot);
  }

  /**
   * Update the snapshot reference
   */
  setSnapshot(snapshot: string | null): void {
    this.snapshot = snapshot;
  }

  /**
   * Build query string with snapshot parameter if set
   */
  private buildQueryString(additionalParams?: Record<string, any>): string {
    const params = new URLSearchParams();
    
    if (this.snapshot) {
      params.append('snapshot', this.snapshot);
    }
    
    if (additionalParams) {
      Object.entries(additionalParams).forEach(([key, value]) => {
        if (value !== undefined && value !== null) {
          params.append(key, String(value));
        }
      });
    }
    
    const queryString = params.toString();
    return queryString ? `?${queryString}` : '';
  }

  /**
   * Remove adapter:// prefix from path if present
   */
  private cleanPath(path: string): string {
    const prefix = `${this.adapter}://`;
    return path.startsWith(prefix) ? path.slice(prefix.length) : path;
  }

  /**
   * Build v2 API URL: /storages/{adapter}/nodes/{path}
   */
  private buildNodeUrl(path: string): string {
    const cleanedPath = this.cleanPath(path);
    const base = `/storages/${this.adapter}/nodes`;
    const url = cleanedPath ? `${base}/${cleanedPath}` : base;
    return `${url}${this.buildQueryString()}`;
  }

  /**
   * Override list method to use v2 API format
   */
  async list(params?: { path?: string }): Promise<any> {
    const url = this.buildNodeUrl(params?.path || '');
    console.log('Listing nodes at URL:', url);

    // console.log("Config", this.config);
    const fullUrl = `${this.config.baseURL}${url}`;
    const options: RequestInit = { method: 'GET' };
    const response = await fetch(fullUrl, {
      ...options,
      headers: {
        ...this.getHeaders(),
        ...(options.headers as Record<string, string>),
      },
    });

    if (response.status === 404) {
        return {
            dirname: params?.path || '',
            read_only: false,
            storages: [this.adapter],
            files: [
                {
                    basename: '[Directory not found]',
                    extension: "",
                    file_size: 0,
                    last_modified: 0,
                    path: params?.path || '',
                    type: "dir",
                    url: "",
                }
            ],
        }
    }

    return response.json();
  }

  /**
   * Override getPreviewUrl to use v2 API format
   */
  getPreviewUrl(params: { path: string }): string {
    return `${this.baseURL}${this.buildNodeUrl(params.path)}`;
  }

  /**
   * Override getDownloadUrl to use v2 API format
   */
  getDownloadUrl(params: { path: string }): string {
    return `${this.baseURL}${this.buildNodeUrl(params.path)}`;
  }

  /**
   * Override getContent to use v2 API format
   */
  async getContent(params: { path: string }): Promise<{ content: string; mimeType?: string }> {
    const url = this.buildNodeUrl(params.path);
    const fullUrl = `${this.baseURL}${url}`;
    
    // Use fetch directly to get response headers
    const headers = (this as any).getHeaders ? (this as any).getHeaders() : {};
    const response = await fetch(fullUrl, { headers });
    
    if (!response.ok) {
      throw new Error(`Failed to get content: ${response.statusText}`);
    }
    
    const content = await response.text();
    return { 
      content, 
      mimeType: response.headers.get('Content-Type') || undefined 
    };
  }
}
