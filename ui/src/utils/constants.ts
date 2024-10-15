import { CredentialsTabIDs } from "src/domain";

// Literals used more than once
export const ACCESSIBLE_UNTIL = "Accessible until";
export const CLOSE = "Close";
export const CONNECTIONS = "Connections";
export const CREDENTIAL_LINK = "Credential link";
export const CREDENTIALS = "Credentials";
export const DELETE = "Delete";
export const DETAILS = "Details";
export const ERROR_MESSAGE = "Something went wrong";
export const EXPIRATION = "Expiration";
export const IDENTIFIER = "Identifier";
export const IMPORT_SCHEMA = "Import schema";
export const ISSUE_CREDENTIAL = "Issue credential";
export const ISSUE_CREDENTIAL_DIRECT = "Issue credential directly";
export const ISSUE_CREDENTIAL_LINK = "Create credential link";
export const ISSUE_DATE = "Issue date";
export const ISSUED = "Issued";
export const ISSUED_CREDENTIALS = "Issued credentials";
export const ISSUER_STATE = "Issuer state";
export const IDENTITY_ADD_NEW = "Add new identity";
export const IDENTITY_ADD = "Add identity";
export const IDENTITY_DETAILS = "Identity details";
export const IDENTITIES = "Identities";
export const LINKS = "Links";
export const REVOCATION = "Revocation";
export const REVOKE = "Revoke";
export const SAVE = "Save";
export const SCHEMA_HASH = "Schema hash";
export const SCHEMA_TYPE = "Schema type";
export const SCHEMAS = "Schemas";
export const STATUS = "Status";
export const VALUE_REQUIRED = "Value required";
export const NOT_PUBLISHED_STATE = "State not published";
export const FINALIZE_SETUP = "Finalize setup";

// URL params
export const DID_SEARCH_PARAM = "did";
export const QUERY_SEARCH_PARAM = "query";
export const SCHEMA_SEARCH_PARAM = "schema";
export const STATUS_SEARCH_PARAM = "status";
export const PAGINATION_PAGE_PARAM = "page";
export const PAGINATION_MAX_RESULTS_PARAM = "max_results";
export const SORT_PARAM = "sort";
export const IDENTIFIER_SEARCH_PARAM = "identifier";

export const DEFAULT_PAGINATION_PAGE = 1;
export const DEFAULT_PAGINATION_MAX_RESULTS = 10;
export const DEFAULT_PAGINATION_TOTAL = 0;

export const API_VERSION = "v2";

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

export const FEEDBACK_URL = "https://forms.gle/W8xuqY3UjPnY5Nj16";

export const IMAGE_PLACEHOLDER_PATH = "/images/image-preview.png";

export const POLLING_INTERVAL = 10000;

export const ROOT_PATH = "/";

export const SIDER_WIDTH = 320;

export const TOAST_NOTIFICATION_TIMEOUT = 6;

export const DOCS_URL = "https://docs.privado.id";

export const WALLET_APP_STORE_URL = "https://apps.apple.com/us/app/polygon-id/id1629870183";

export const WALLET_PLAY_STORE_URL =
  "https://play.google.com/store/apps/details?id=com.polygonid.wallet";

export const IPFS_PUBLIC_GATEWAY_CHECKER_URL = "https://ipfs.github.io/public-gateway-checker";

export const IPFS_CUSTOM_GATEWAY_KEY = "ipfsGatewayUrl";

export const URL_FIELD_ERROR_MESSAGE =
  "Must be a valid URL that includes a scheme such as https://";

export const IDENTIFIER_LOCAL_STORAGE_KEY = "identifier";
