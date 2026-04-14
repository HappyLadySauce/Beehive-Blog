import axios, { AxiosError, InternalAxiosRequestConfig, AxiosResponse } from 'axios';
import { useAuthStore } from '../store/authStore';

// Create an axios instance
const request = axios.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || '',
  timeout: 10000,
});

// Flag to indicate if a refresh token request is in progress
let isRefreshing = false;
// Queue to hold pending requests while token is being refreshed
let requestsQueue: ((token: string) => void)[] = [];

// Request interceptor
request.interceptors.request.use(
  (config: InternalAxiosRequestConfig) => {
    const token = useAuthStore.getState().token;
    if (token && config.headers) {
      config.headers.Authorization = `Bearer ${token}`;
    }
    return config;
  },
  (error: AxiosError) => {
    return Promise.reject(error);
  }
);

// Response interceptor
request.interceptors.response.use(
  (response: AxiosResponse) => {
    // You can handle custom response codes here if needed
    return response.data;
  },
  async (error: AxiosError) => {
    const originalRequest = error.config as InternalAxiosRequestConfig & { _retry?: boolean };

    if (error.response?.status === 401 && !originalRequest._retry) {
      if (isRefreshing) {
        // If already refreshing, queue the request
        return new Promise((resolve) => {
          requestsQueue.push((token: string) => {
            if (originalRequest.headers) {
              originalRequest.headers.Authorization = `Bearer ${token}`;
            }
            resolve(request(originalRequest));
          });
        });
      }

      originalRequest._retry = true;
      isRefreshing = true;

      try {
        const refreshToken = useAuthStore.getState().refreshToken;
        if (!refreshToken) {
          throw new Error('No refresh token available');
        }

        // Call refresh token API
        const res = await axios.post(`${request.defaults.baseURL}/api/v1/auth/refresh`, {
          refreshToken,
        });

        const { token, refreshToken: newRefreshToken } = res.data.data;
        
        // Update store
        useAuthStore.getState().setTokens(token, newRefreshToken);

        // Process queued requests
        requestsQueue.forEach((cb) => cb(token));
        requestsQueue = [];

        // Retry the original request
        if (originalRequest.headers) {
          originalRequest.headers.Authorization = `Bearer ${token}`;
        }
        return request(originalRequest);
      } catch (refreshError) {
        // If refresh fails, clear tokens and redirect to login
        useAuthStore.getState().logout();
        requestsQueue = [];
        window.location.href = '/login';
        return Promise.reject(refreshError);
      } finally {
        isRefreshing = false;
      }
    }

    return Promise.reject(error);
  }
);

export default request;
