export interface Deployment {
  chainName: string;
  tokenSymbol: string;
  blockchainID: string;
  subnetID?: string;
  rpcURL: string;
  chainID: number;
  network: string;
  deployedAt: string;
  status?: string;
}

export type ChainState = "online" | "offline" | "pending";

export interface ChainStatus {
  chainName: string;
  status: ChainState;
  blockNumber?: string;
  message?: string;
  peerCount?: number;
  gasPrice?: string;
  checkedAt: string;
}

export interface Validator {
  nodeID: string;
  weight: number;
  status: string;
  uptime?: string;
  validationID?: string;
}

export interface TeleporterMessage {
  messageID: string;
  sourceChain: string;
  destChain: string;
  sender: string;
  receiver: string;
  nonce: number;
  status: string;
  timestamp: string;
}

export interface TeleporterInfo {
  messengerAddress: string;
  registryAddress: string;
  relayerStatus: string;
  messages: TeleporterMessage[];
}

export interface ApiError {
  message: string;
}