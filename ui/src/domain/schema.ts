export interface Schema {
  bigInt: string;
  createdAt: Date;
  hash: string;
  id: string;
  type: string;
  url: string;
}

export type SchemaAttribute = {
  description?: string;
  name: string;
  technicalName: string;
} & (
  | {
      type: "number" | "boolean" | "date";
    }
  | {
      type: "singlechoice";
      values: {
        name: string;
        value: number;
      }[];
    }
);
