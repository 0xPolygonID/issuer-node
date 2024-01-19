import { useCallback, useEffect, useState } from "react";
import { useParams } from "react-router-dom";

import { getIssuedQRCodes } from "src/adapters/api/credentials";
import { CredentialQR } from "src/components/credentials/CredentialQR";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { useEnvContext } from "src/contexts/Env";
import { AppError, IssuedQRCode } from "src/domain";
import { AsyncTask, hasAsyncTaskFailed, isAsyncTaskDataAvailable } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";

export function CredentialIssuedQR() {
  const env = useEnvContext();

  const [issuedQRCodes, setIssuedQRCodes] = useState<
    AsyncTask<[IssuedQRCode, IssuedQRCode], AppError>
  >({
    status: "pending",
  });

  const { credentialID } = useParams();

  const createCredentialQR = useCallback(
    async (signal: AbortSignal) => {
      if (credentialID) {
        setIssuedQRCodes({ status: "loading" });

        const response = await getIssuedQRCodes({ credentialID, env, signal });

        if (response.success) {
          setIssuedQRCodes({ data: response.data, status: "successful" });
        } else {
          if (!isAbortedError(response.error)) {
            setIssuedQRCodes({ error: response.error, status: "failed" });
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
    setIssuedQRCodes({ status: "pending" });
  };

  if (hasAsyncTaskFailed(issuedQRCodes)) {
    return (
      <ErrorResult
        error={issuedQRCodes.error.message}
        labelRetry="Start again"
        onRetry={onStartAgain}
      />
    );
  }

  if (!isAsyncTaskDataAvailable(issuedQRCodes)) {
    return <LoadingResult />;
  }

  const [issuedQRCodeLink, issuedQRCodeRaw] = issuedQRCodes.data;
  return (
    <CredentialQR
      qrCodeBase64={`iden3comm://?i_m=${btoa(issuedQRCodeRaw.qrCode)}`}
      qrCodeLink={issuedQRCodeLink.qrCode}
      qrCodeRaw={issuedQRCodeRaw.qrCode}
      schemaType={issuedQRCodeLink.schemaType}
      subTitle="Scan the QR code with your Polygon ID wallet to add the credential."
    />
  );
}
