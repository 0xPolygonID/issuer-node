import { CredentialsTabIDs } from "src/domain";

// Literals used more than once
export const ACCESSIBLE_UNTIL = "Accessible until";
export const CLOSE = "Close";
export const CONNECTIONS = "Connections";
export const CREDENTIAL_LINK = "Credential link";
export const CREDENTIALS = "Credentials";
export const DATE_VALIDITY_MESSAGE = "Valid date required";
export const TIME_VALIDITY_MESSAGE = "Valid time required";
export const ERROR_MESSAGE = "Something went wrong";
export const EXPIRATION = "Expiration";
export const IDENTIFIER = "Identifier";
export const IMPORT_SCHEMA = "Import schema";
export const ISSUE_CREDENTIAL = "Issue credential";
export const ISSUE_DATE = "Issue date";
export const ISSUED = "Issued";
export const ISSUER_STATE = "Issuer state";
export const LINKS = "Links";
export const REVOCATION = "Revocation";
export const SCHEMA_HASH = "Schema hash";
export const SCHEMA_TYPE = "Schema type";
export const SCHEMAS = "Schemas";
export const STATUS = "Status";

// URL params
export const DID_SEARCH_PARAM = "did";
export const QUERY_SEARCH_PARAM = "query";
export const STATUS_SEARCH_PARAM = "status";

export const API_VERSION = "v1";

type CredentialsTab = { id: CredentialsTabIDs; tabID: string; title: string };

export const CREDENTIALS_TABS: [CredentialsTab, CredentialsTab] = [
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

export const DOTS_DROPDOWN_WIDTH = 60;

export const FEEDBACK_URL = "https://forms.gle/ckDgvw1e9yZJBNfH6";

export const IMAGE_PLACEHOLDER_PATH = "/images/image-preview.png";

export const POLLING_INTERVAL = 10000;

export const ROOT_PATH = "/";

export const SIDER_WIDTH = 320;

export const TOAST_NOTIFICATION_TIMEOUT = 6;

export const TUTORIALS_URL = "https://0xpolygonid.github.io/tutorials";

export const WALLET_APP_STORE_URL = "https://apps.apple.com/us/app/polygon-id/id1629870183";

export const WALLET_PLAY_STORE_URL =
  "https://play.google.com/store/apps/details?id=com.polygonid.wallet";
