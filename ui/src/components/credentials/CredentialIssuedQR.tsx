import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";

import { getIssuedQRCode } from "src/adapters/api/credentials";
import { CredentialQR } from "src/components/credentials/CredentialQR";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { useEnvContext } from "src/contexts/Env";
import { AppError, IssuedQRCode } from "src/domain";
import { AsyncTask, hasAsyncTaskFailed, isAsyncTaskDataAvailable } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";

export function CredentialIssuedQR() {
  const env = useEnvContext();

  const [issuedQRCode, setIssuedQRCode] = useState<AsyncTask<IssuedQRCode, AppError>>({
    status: "pending",
  });

  const { credentialID } = useParams();

  const createCredentialQR = useCallback(
    async (signal: AbortSignal) => {
      if (credentialID) {
        setIssuedQRCode({ status: "loading" });

        const response = await getIssuedQRCode({ credentialID, env, signal });

        if (response.success) {
          setIssuedQRCode({ data: response.data, status: "successful" });
        } else {
          if (!isAbortedError(response.error)) {
            setIssuedQRCode({ error: response.error, status: "failed" });
          }
        }
      }
    },
    [credentialID, env]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(createCredentialQR);

    return aborter;
  }, [createCredentialQR]);

  const onStartAgain = () => {
    makeRequestAbortable(createCredentialQR);
    setIssuedQRCode({ status: "pending" });
  };

  if (hasAsyncTaskFailed(issuedQRCode)) {
    return (
      <ErrorResult
        error={issuedQRCode.error.message}
        labelRetry="Start again"
        onRetry={onStartAgain}
      />
    );
  }

  if (!isAsyncTaskDataAvailable(issuedQRCode)) {
    return <LoadingResult />;
  }

  return (
    <CredentialQR
      qrCode={issuedQRCode.data.qrCodeLink}
      schemaType={issuedQRCode.data.schemaType}
      subTitle="Scan the QR code with your Polygon ID wallet to add the credential."
    />
  );
}
