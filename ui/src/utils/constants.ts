import { CredentialsTabIDs } from "src/domain";
import { RequestsTabIDs } from "src/domain/request";

// Literals used more than once
export const ACCESSIBLE_UNTIL = "Accessible until";
export const CLOSE = "Close";
export const CONNECTIONS = "Connections";
export const CREDENTIAL_LINK = "Credential link";
export const CREDENTIALS = "Credentials";
export const DELETE = "Delete";
export const DETAILS = "Details";
export const APPROVE1 = "Approve 1";
export const APPROVE2 = "Approve 2";
export const ERROR_MESSAGE = "Something went wrong";
export const EXPIRATION = "Expiration";
export const EXPIRED = "Expired";
export const REVOKE_DATE = "Revoke Date";
export const IDENTIFIER = "Identifier";
export const IMPORT_SCHEMA = "Import schema";
export const ISSUE_CREDENTIAL = "Issue credential";
export const ISSUE_REQUEST = "My Requests";
export const REQUEST_FOR_VC_CREDS = "Request for VC verification";
export const REQUEST_FOR_VC = "Request for VC";
export const ISSUE_CREDENTIAL_DIRECT = "Issue credential directly";
export const ISSUE_CREDENTIAL_LINK = "Create credential link";
export const ISSUE_DATE = "Issue date";
export const ISSUED = "Issued";
export const ISSUED_CREDENTIALS = "Issued credentials";
export const ISSUER_STATE = "Issuer state";
export const LINKS = "Links";
export const REVOCATION = "Revocation";
export const REVOKE = "Revoke";
export const SCHEMA_HASH = "Schema hash";
export const SCHEMA_TYPE = "Schema type";
export const SCHEMAS = "Schemas";
export const NOTIFICATION = "Notifications";
export const REQUEST = "Request";
export const REQUESTS = "Requests";
export const STATUS = "Status";
export const VALUE_REQUIRED = "Value required";
export const REQUEST_DATE = "Request date";

// URL params
export const DID_SEARCH_PARAM = "did";
export const QUERY_SEARCH_PARAM = "query";
export const SCHEMA_SEARCH_PARAM = "schema";
export const STATUS_SEARCH_PARAM = "status";

export const API_VERSION = "v1";

type CredentialsTab = { id: CredentialsTabIDs; tabID: string; title: string };

export const CREDENTIALS_TABS: CredentialsTab[] = [
  {
    id: "issued",
    tabID: "issued",
    title: ISSUED,
  },
];

type RequestsTab = { id: RequestsTabIDs; tabID: string; title: string };
export const REQUEST_TABS: RequestsTab[] = [
  {
    id: "Request",
    tabID: "Request",
    title: REQUEST,
  },
];

export const DEBOUNCE_INPUT_TIMEOUT = 500;

export const DOTS_DROPDOWN_WIDTH = 60;

export const FEEDBACK_URL = "https://forms.gle/W8xuqY3UjPnY5Nj16";

export const IMAGE_PLACEHOLDER_PATH = "/images/image-preview.png";

export const POLLING_INTERVAL = 10000;

export const ROOT_PATH = "/";

export const SIDER_WIDTH = 320;

export const TOAST_NOTIFICATION_TIMEOUT = 6;

export const TUTORIALS_URL = "https://0xpolygonid.github.io/tutorials";

export const WALLET_APP_STORE_URL = "https://apps.apple.com/us/app/polygon-id/id1629870183";

export const WALLET_PLAY_STORE_URL =
  "https://play.google.com/store/apps/details?id=com.polygonid.wallet";
