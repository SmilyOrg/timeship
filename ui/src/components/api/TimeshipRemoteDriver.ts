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

  constructor(config: any) {
    super(config);
    this.adapter = config.adapter || 'local';
    this.baseURL = config.baseURL || '';
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
    return cleanedPath ? `${base}/${cleanedPath}` : base;
  }

  /**
   * Override list method to use v2 API format
   */
  async list(params?: { path?: string }): Promise<any> {
    const url = this.buildNodeUrl(params?.path || '');
    // Access parent's private request method via bracket notation
    return await (this as any).request(url, { method: 'GET' });
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
