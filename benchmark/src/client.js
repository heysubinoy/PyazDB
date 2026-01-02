/**
 * PyazDB API Client
 * Handles communication with the PyazDB HTTP API
 */
export class PyazDBClient {
  constructor(endpoint) {
    this.endpoint = endpoint.replace(/\/$/, '');
    this.useProxy = true; // Use Vite proxy by default
  }

  setEndpoint(endpoint) {
    this.endpoint = endpoint.replace(/\/$/, '');
  }

  /**
   * Get the actual URL to use (proxy or direct)
   */
  getUrl(path) {
    if (this.useProxy) {
      return `/pyazdb${path}`;
    }
    return `${this.endpoint}${path}`;
  }

  /**
   * PUT operation - Set a key-value pair
   * @param {string} key 
   * @param {string} value 
   * @returns {Promise<{success: boolean, latency: number, error?: string}>}
   */
  async put(key, value) {
    const start = performance.now();
    try {
      const response = await fetch(this.getUrl('/set'), {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ key, value }),
      });
      
      const latency = performance.now() - start;
      
      if (response.status === 204 || response.ok) {
        return { success: true, latency };
      }
      
      const errorText = await response.text();
      return { success: false, latency, error: errorText || `HTTP ${response.status}` };
    } catch (error) {
      const latency = performance.now() - start;
      return { success: false, latency, error: error.message };
    }
  }

  /**
   * GET operation - Retrieve a value by key
   * @param {string} key 
   * @returns {Promise<{success: boolean, latency: number, value?: string, error?: string}>}
   */
  async get(key) {
    const start = performance.now();
    try {
      const response = await fetch(this.getUrl(`/get?key=${encodeURIComponent(key)}`), {
        method: 'GET',
      });
      
      const latency = performance.now() - start;
      
      if (response.ok) {
        const value = await response.text();
        return { success: true, latency, value };
      }
      
      if (response.status === 404) {
        return { success: false, latency, error: 'Key not found' };
      }
      
      const errorText = await response.text();
      return { success: false, latency, error: errorText || `HTTP ${response.status}` };
    } catch (error) {
      const latency = performance.now() - start;
      return { success: false, latency, error: error.message };
    }
  }

  /**
   * DELETE operation - Remove a key
   * @param {string} key 
   * @returns {Promise<{success: boolean, latency: number, error?: string}>}
   */
  async delete(key) {
    const start = performance.now();
    try {
      const response = await fetch(this.getUrl('/delete'), {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ key }),
      });
      
      const latency = performance.now() - start;
      
      if (response.status === 204 || response.ok) {
        return { success: true, latency };
      }
      
      const errorText = await response.text();
      return { success: false, latency, error: errorText || `HTTP ${response.status}` };
    } catch (error) {
      const latency = performance.now() - start;
      return { success: false, latency, error: error.message };
    }
  }

  /**
   * Health check - verify connection to the endpoint
   * @returns {Promise<boolean>}
   */
  async healthCheck() {
    try {
      // Try a simple GET with a non-existent key
      const response = await fetch(this.getUrl('/get?key=__health_check__'), {
        method: 'GET',
        signal: AbortSignal.timeout(5000),
      });
      // 404 means the server is responding (key just doesn't exist)
      return response.ok || response.status === 404;
    } catch {
      return false;
    }
  }
}
