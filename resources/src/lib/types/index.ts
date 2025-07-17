export type FlashMessageType = 'success' | 'error' | 'info' | 'warning';

export interface FlashMessage {
  type: FlashMessageType;
  message: string;
}
