import { z } from "zod";

import { getStrictParser } from "src/adapters/parsers";
import { Env } from "src/domain";
import { IMAGE_PLACEHOLDER_PATH } from "src/utils/constants";

export type EnvInput = {
  VITE_API_PASSWORD: string;
  VITE_API_URL: string;
  VITE_API_USERNAME: string;
  VITE_BLOCK_EXPLORER_URL: string;
  VITE_BUILD_TAG?: string;
  VITE_IPFS_GATEWAY_URL: string;
  VITE_ISSUER_DID: string;
  VITE_ISSUER_LOGO?: string;
  VITE_ISSUER_NAME: string;
  VITE_WARNING_MESSAGE?: string;
};

export const envParser = getStrictParser<EnvInput, Env>()(
  z
    .object({
      VITE_API_PASSWORD: z.string().min(1),
      VITE_API_URL: z.string().url(),
      VITE_API_USERNAME: z.string().min(1),
      VITE_BLOCK_EXPLORER_URL: z.string().url(),
      VITE_BUILD_TAG: z.string().optional(),
      VITE_IPFS_GATEWAY_URL: z.string(),
      VITE_ISSUER_DID: z.string(),
      VITE_ISSUER_LOGO: z.string().optional(),
      VITE_ISSUER_NAME: z.string().min(1),
      VITE_WARNING_MESSAGE: z.string().optional(),
    })
    .transform(
      ({
        VITE_API_PASSWORD,
        VITE_API_URL,
        VITE_API_USERNAME,
        VITE_BLOCK_EXPLORER_URL,
        VITE_BUILD_TAG,
        VITE_IPFS_GATEWAY_URL,
        VITE_ISSUER_DID,
        VITE_ISSUER_LOGO,
        VITE_ISSUER_NAME,
        VITE_WARNING_MESSAGE,
      }): Env => ({
        api: {
          password: VITE_API_PASSWORD,
          url: VITE_API_URL,
          username: VITE_API_USERNAME,
        },
        blockExplorerUrl: VITE_BLOCK_EXPLORER_URL,
        buildTag: VITE_BUILD_TAG,
        ipfsGatewayUrl: VITE_IPFS_GATEWAY_URL,
        issuer: {
          did: VITE_ISSUER_DID,
          logo: VITE_ISSUER_LOGO || IMAGE_PLACEHOLDER_PATH,
          name: VITE_ISSUER_NAME,
        },
        warningMessage: VITE_WARNING_MESSAGE,
      })
    )
);
