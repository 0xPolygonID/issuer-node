import { Button, Card, Form, Input, Row } from "antd";
import axios from "axios";
import { useState } from "react";
import { generatePath } from "react-router-dom";
import { ZodError } from "zod";

import { schemaParser } from "src/adapters/parsers/schemas";
import { ErrorResult } from "src/components/schemas/ErrorResult";
import { SchemaTree } from "src/components/schemas/SchemaTree";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { Schema } from "src/domain";
import { ROUTES } from "src/routes";
import { processZodError } from "src/utils/adapters";
import { CARD_WIDTH, SCHEMAS_TABS } from "src/utils/constants";
import { AsyncTask } from "src/utils/types";

export function ImportSchema() {
  const [schemaUrl, setSchemaUrl] = useState<string>();
  const [schema, setSchema] = useState<AsyncTask<Schema, string | ZodError>>({ status: "pending" });

  const fetchSchemaFromUrl = (url: string): void => {
    axios({
      method: "GET",
      url,
    })
      .then((response) => {
        const parsedSchema = schemaParser.safeParse(response.data);

        if (parsedSchema.success) {
          setSchema({ data: parsedSchema.data, status: "successful" });
        } else {
          setSchema({ error: parsedSchema.error, status: "failed" });
        }
      })
      .catch((error) => {
        setSchema({
          error:
            error instanceof ZodError
              ? error
              : error instanceof Error
              ? error.message
              : "Unknown error",
          status: "failed",
        });
      });
  };

  const loadSchema = () => {
    if (schemaUrl && (schemaUrl.startsWith("http://") || schemaUrl.startsWith("https://"))) {
      setSchema({ status: "loading" });
      fetchSchemaFromUrl(schemaUrl);
    }
  };

  return (
    <SiderLayoutContent
      backButtonLink={generatePath(ROUTES.schemas.path, { tabID: SCHEMAS_TABS[0].tabID })}
      description="Preview, import and use verifiable credential schemas."
      showDivider
      title="Import schema"
    >
      <Card style={{ margin: "auto", maxWidth: CARD_WIDTH }} title="Paste JSON schema URL">
        <Form layout="vertical" onFinish={loadSchema}>
          <Form.Item label="JSON schema *">
            <Input.Group className="input-copy-group" compact>
              <Input
                onChange={(event) => setSchemaUrl(event.target.value)}
                placeholder="Enter JSON schema"
                value={schemaUrl}
              />
            </Input.Group>
          </Form.Item>
          <Row justify="end">
            <Button disabled={!schemaUrl} htmlType="submit" type="primary">
              Preview import
            </Button>
          </Row>
        </Form>

        {(() => {
          switch (schema.status) {
            case "pending": {
              return null;
            }
            case "loading": {
              return <LoadingResult />;
            }
            case "failed": {
              return (
                <ErrorResult
                  error={
                    schema.error instanceof ZodError
                      ? [
                          "An error occurred while trying to parse this schema. Please make sure you provide a valid JSON Schema.\n",
                          ...processZodError(schema.error),
                        ].join("\n")
                      : `An error occurred while trying to load this schema. ${schema.error}`
                  }
                />
              );
            }
            case "successful":
            case "reloading": {
              return (
                <SchemaTree schema={schema.data} style={{ background: "#F2F4F7", padding: 24 }} />
              );
            }
          }
        })()}
      </Card>
    </SiderLayoutContent>
  );
}
