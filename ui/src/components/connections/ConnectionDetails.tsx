import { Card, Space, message } from "antd";
import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";

import { Credential, credentialIssue } from "src/adapters/api/credentials";
import { Schema, getSchema } from "src/adapters/api/schemas";
import { issueCredentialFormData } from "src/adapters/parsers/forms";
import { serializeCredentialForm } from "src/adapters/parsers/serializers";
import {
  AttributeValues,
  IssueCredentialForm,
} from "src/components/credentials/IssueCredentialForm";
import { SelectSchema } from "src/components/credentials/SelectSchema";
import { IssuanceMethod, SetIssuanceMethod } from "src/components/credentials/SetIssuanceMethod";
import { Summary } from "src/components/credentials/Summary";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/env";
import { APIError, processZodError } from "src/utils/adapters";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { ISSUE_CREDENTIAL } from "src/utils/constants";
import { AsyncTask, isAsyncTaskDataAvailable } from "src/utils/types";

export function ConnectionDetails() {
  //   const env = useEnvContext();

  const { connectionID } = useParams();

  return (
    <SiderLayoutContent
      description="View connection information, credential attribute data. Revoke and delete issued credentials."
      showBackButton
      showDivider
      title="Connection details"
    >
      <Card></Card>
      <Card></Card>
    </SiderLayoutContent>
  );
}
