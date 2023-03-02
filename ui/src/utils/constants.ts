import { TabsSchemasIDs } from "src/domain";

export const FORM_LABEL = {
  ARCHIVED_SCHEMAS: "Archived schemas",
  ATTRIBUTES: "Attributes",
  CLAIM_AVAILABILITY: "No. of claims left",
  CLAIM_EXPIRATION: "Claim expiration date",
  CLAIM_LINK: "Claim link",
  CREATION_DATE: "Creation date",
  EXPIRATION_DATE: "Expiration date",
  LINK_VALIDITY: "Link valid until",
  MY_SCHEMAS: "My schemas",
  SCHEMA_HASH: "Schema hash",
  SCHEMA_ID: "Schema ID",
  SCHEMA_NAME: "Schema name",
  SCHEMA_TYPE: "Schema type",
};

export const ACTIVE_SEARCH_PARAM = "active";

export const API_URL = import.meta.env.VITE_API;

export const AUTH_CONTEXT_NOT_READY_MESSAGE = "Auth Context is not ready yet";

export const AUTH_TOKEN_KEY = "authToken";

export const CARD_ELLIPSIS_MAXIMUM_WIDTH = "66%";

export const CARD_WIDTH = 720;

export const CLAIM_ID_SEARCH_PARAM = "claimID";

export const SCHEMAS_TABS: { id: TabsSchemasIDs; tabID: string; title: string }[] = [
  {
    id: "mySchemas",
    tabID: "my-schemas",
    title: FORM_LABEL.MY_SCHEMAS,
  },
  {
    id: "archivedSchemas",
    tabID: "archived-schemas",
    title: FORM_LABEL.ARCHIVED_SCHEMAS,
  },
];

export const CONTENT_WIDTH = 585;

export const COPYRIGHT_URL = "https://polygonid.com";

export const DATE_VALIDITY_MESSAGE = "Valid date required";

export const DEBOUNCE_INPUT_TIMEOUT = 500;

export const ROOT_PATH = "/";

export const DETAILS_MAXIMUM_WIDTH = 216;

export const ERROR_MESSAGE = "Something went wrong";

export const FEEDBACK_URL = "https://forms.gle/ckDgvw1e9yZJBNfH6";

export const IMAGE_PLACEHOLDER_PATH = "/images/image-preview.png";

export const QR_CODE_POLLING_INTERVAL = 10000;

export const QUERY_SEARCH_PARAM = "query";

export const SCHEMA_FORM_EXTRA_ALPHA_MESSAGE =
  "Only alphanumeric characters allowed and no spaces. Not seen by end users.";

export const SCHEMA_FORM_HELP_ALPHA_MESSAGE = "Only alphanumeric characters allowed.";

export const SCHEMA_FORM_HELP_NUMERIC_MESSAGE = "Only numbers allowed.";

export const SCHEMA_FORM_HELP_REQUIRED_MESSAGE = "Required field.";

export const SCHEMA_ID_SEARCH_PARAM = "schemaID";

export const SCHEMA_KEY_MAX_LENGTH = 32;

export const SIDER_WIDTH = 320;

export const TOAST_NOTIFICATION_TIMEOUT = 6;

export const TUTORIALS_URL = "https://0xpolygonid.github.io/tutorials";

export const VALID_SEARCH_PARAM = "valid";

export const WALLET_APP_STORE_URL = "https://apps.apple.com/us/app/polygon-id/id1629870183";

export const WALLET_PLAY_STORE_URL =
  "https://play.google.com/store/apps/details?id=com.polygonid.wallet";
