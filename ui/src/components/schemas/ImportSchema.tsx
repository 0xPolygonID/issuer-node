import { useState } from "react";

import { FormData, ImportSchemaForm } from "src/components/schemas/ImportSchemaForm";
import { ImportSchemaPreview } from "src/components/schemas/ImportSchemaPreview";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
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
  const [step, setStep] = useState<Step>({ type: "form" });

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
          rawJsonLdContext={step.formData.rawJsonLdContext}
          rawJsonSchema={step.formData.rawJsonSchema}
          schema={step.formData.schema}
          url={step.formData.schemaUrl}
        />
      )}
    </SiderLayoutContent>
  );
}
