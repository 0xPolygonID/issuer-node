import { env } from "src/adapters/parsers/env";
import { TabsCredentialsIDs } from "src/domain";

// Literals used more than once
export const ACCESSIBLE_UNTIL = "Accessible until";
export const AUTH_CONTEXT_NOT_READY_MESSAGE = "Auth Context is not ready yet";
export const CREDENTIAL_LINK = "Credential link";
export const CREDENTIALS = "Credentials";
export const DATE_VALIDITY_MESSAGE = "Valid date required";
export const ERROR_MESSAGE = "Something went wrong";
export const IMPORT_SCHEMA = "Import schema";
export const ISSUE_CREDENTIAL = "Issue credential";
export const ISSUED = "Issued";
export const LINKS = "Links";
export const SCHEMA_HASH = "Schema hash";
export const SCHEMA_TYPE = "Schema type";
export const SCHEMAS = "Schemas";

export const API_AUTH = `Basic ${env.api.username}:${env.api.password}`;

export const AUTH_TOKEN_KEY = "authToken";

export const CARD_ELLIPSIS_MAXIMUM_WIDTH = "66%";

export const CARD_WIDTH = 720;

export const CONTENT_WIDTH = 585;

export const CREDENTIALS_TABS: { id: TabsCredentialsIDs; tabID: string; title: string }[] = [
  {
    id: "issued",
    tabID: "issued",
    title: ISSUED,
  },
  {
    id: "links",
    tabID: "links",
    title: LINKS,
  },
];

export const DEBOUNCE_INPUT_TIMEOUT = 500;

export const FEEDBACK_URL = "https://forms.gle/ckDgvw1e9yZJBNfH6";

export const IMAGE_PLACEHOLDER_PATH = "/images/image-preview.png";

export const QR_CODE_POLLING_INTERVAL = 10000;

export const QUERY_SEARCH_PARAM = "query";

export const ROOT_PATH = "/";

export const SIDER_WIDTH = 320;

export const TOAST_NOTIFICATION_TIMEOUT = 6;

export const TUTORIALS_URL = "https://0xpolygonid.github.io/tutorials";

export const WALLET_APP_STORE_URL = "https://apps.apple.com/us/app/polygon-id/id1629870183";

export const WALLET_PLAY_STORE_URL =
  "https://play.google.com/store/apps/details?id=com.polygonid.wallet";
