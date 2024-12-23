import { z } from "zod";

import { getStrictParser } from "src/adapters/parsers";
import { Env } from "src/domain";
import { IMAGE_PLACEHOLDER_PATH } from "src/utils/constants";

export type EnvInput = {
  VITE_API_PASSWORD: string;
  VITE_API_URL: string;
  VITE_API_USERNAME: string;
  VITE_BASE_URL?: string;
  VITE_BUILD_TAG?: string;
  VITE_DISPLAY_METHOD_BUILDER_URL: string;
  VITE_IPFS_GATEWAY_URL: string;
  VITE_ISSUER_LOGO?: string;
  VITE_ISSUER_NAME: string;
  VITE_SCHEMA_EXPLORER_AND_BUILDER_URL?: string;
  VITE_WARNING_MESSAGE?: string;
};

export const envParser = getStrictParser<EnvInput, Env>()(
  z
    .object({
      VITE_API_PASSWORD: z.string().min(1),
      VITE_API_URL: z.string().url(),
      VITE_API_USERNAME: z.string().min(1),
      VITE_BASE_URL: z.string().optional(),
      VITE_BUILD_TAG: z.string().optional(),
      VITE_DISPLAY_METHOD_BUILDER_URL: z.string(),
      VITE_IPFS_GATEWAY_URL: z.string().url(),
      VITE_ISSUER_LOGO: z.string().optional(),
      VITE_ISSUER_NAME: z.string().min(1),
      VITE_SCHEMA_EXPLORER_AND_BUILDER_URL: z
        .union([z.string().url(), z.literal("")])
        .transform((value) => value || undefined)
        .optional(),
      VITE_WARNING_MESSAGE: z.string().optional(),
    })
    .transform(
      ({
        VITE_API_PASSWORD,
        VITE_API_URL,
        VITE_API_USERNAME,
        VITE_BASE_URL,
        VITE_BUILD_TAG,
        VITE_DISPLAY_METHOD_BUILDER_URL,
        VITE_IPFS_GATEWAY_URL,
        VITE_ISSUER_LOGO,
        VITE_ISSUER_NAME,
        VITE_SCHEMA_EXPLORER_AND_BUILDER_URL,
        VITE_WARNING_MESSAGE,
      }): Env => ({
        api: {
          password: VITE_API_PASSWORD,
          url: VITE_API_URL,
          username: VITE_API_USERNAME,
        },
        baseUrl: VITE_BASE_URL,
        buildTag: VITE_BUILD_TAG,
        displayMethodBuilderUrl: VITE_DISPLAY_METHOD_BUILDER_URL,
        ipfsGatewayUrl: VITE_IPFS_GATEWAY_URL,
        issuer: {
          logo: VITE_ISSUER_LOGO || IMAGE_PLACEHOLDER_PATH,
          name: VITE_ISSUER_NAME,
        },
        schemaExplorerAndBuilderUrl: VITE_SCHEMA_EXPLORER_AND_BUILDER_URL,
        warningMessage: VITE_WARNING_MESSAGE,
      })
    )
);
