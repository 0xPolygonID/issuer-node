declare module "ajv-formats-draft2019" {
  import { Ajv } from "ajv";
  import { Ajv2020 } from "ajv/dist/2020";
  type AnyAjv = Ajv | Ajv2020;
  const apply: (ajv: AnyAjv) => AnyAjv;
  // eslint-disable-next-line import/no-default-export
  export default apply;
}
