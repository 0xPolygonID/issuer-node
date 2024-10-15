import { App } from "antd";
import { useState } from "react";
import { useNavigate } from "react-router-dom";

import { importSchema } from "src/adapters/api/schemas";
import { FormData, ImportSchemaForm } from "src/components/schemas/ImportSchemaForm";
import { ImportSchemaPreview } from "src/components/schemas/ImportSchemaPreview";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { ROUTES } from "src/routes";
import { IMPORT_SCHEMA } from "src/utils/constants";

type Step =
  | {
      formData?: FormData;
      type: "form";
    }
  | {
      formData: FormData;
      type: "preview";
    };

export function ImportSchema() {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();
  const navigate = useNavigate();

  const { message } = App.useApp();

  const [step, setStep] = useState<Step>({ type: "form" });

  const onSchemaImport = ({
    jsonLdType,
    jsonSchema: {
      jsonSchemaProps: {
        $metadata: { version },
      },
      schema: { description, title },
    },
    schemaUrl,
  }: FormData) =>
    void importSchema({
      description,
      env,
      identifier,
      jsonLdType,
      schemaUrl,
      title,
      version,
    }).then((response) => {
      if (response.success) {
        navigate(ROUTES.schemas.path);

        void message.success("Schema successfully imported");
      } else {
        void message.error(response.error.message);
      }
    });

  return (
    <SiderLayoutContent
      description="Preview, import and use verifiable credential schemas."
      showBackButton
      showDivider
      title={IMPORT_SCHEMA}
    >
      {step.type === "form" ? (
        <ImportSchemaForm
          initialFormData={step.formData}
          onFinish={(formData) => {
            setStep({
              formData,
              type: "preview",
            });
          }}
        />
      ) : (
        <ImportSchemaPreview
          jsonLdContextObject={step.formData.jsonLdContextObject}
          jsonLdType={step.formData.jsonLdType}
          jsonSchema={step.formData.jsonSchema}
          jsonSchemaObject={step.formData.jsonSchemaObject}
          onBack={() => {
            setStep({ formData: step.formData, type: "form" });
          }}
          onImport={() => {
            onSchemaImport(step.formData);
          }}
          url={step.formData.schemaUrl}
        />
      )}
    </SiderLayoutContent>
  );
}
