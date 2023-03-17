import { message } from "antd";
import { useState } from "react";
import { useNavigate } from "react-router-dom";

import { importSchema } from "src/adapters/api/schemas";
import { FormData, ImportSchemaForm } from "src/components/schemas/ImportSchemaForm";
import { ImportSchemaPreview } from "src/components/schemas/ImportSchemaPreview";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
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
  const navigate = useNavigate();
  const [step, setStep] = useState<Step>({ type: "form" });

  const onSchemaImport = ({ jsonLdType, schemaUrl: url }: FormData) => {
    void importSchema({ jsonLdType, url }).then((result) => {
      if (result.isSuccessful) {
        void message.info("Schema successfully imported");
        navigate(ROUTES.schemas.path);
      } else {
        void message.error(result.error.message);
      }
    });
  };

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
          jsonLdType={step.formData.jsonLdType}
          onBack={() => {
            setStep({ formData: step.formData, type: "form" });
          }}
          onImport={() => {
            onSchemaImport(step.formData);
          }}
          rawJsonLdContext={step.formData.rawJsonLdContext}
          rawJsonSchema={step.formData.rawJsonSchema}
          schema={step.formData.schema}
          url={step.formData.schemaUrl}
        />
      )}
    </SiderLayoutContent>
  );
}
