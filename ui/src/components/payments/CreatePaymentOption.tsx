import { App, Card } from "antd";
import { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

import { createPaymentOption, getPaymentConfigurations } from "src/adapters/api/payments";
import { notifyParseError } from "src/adapters/parsers";
import { PaymentOptionFormData, paymentOptionFormParser } from "src/adapters/parsers/view";
import { PaymentOptionForm } from "src/components/payments/PaymentOptionForm";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { AppError, PaymentConfigurations } from "src/domain";
import { ROUTES } from "src/routes";
import {
  AsyncTask,
  hasAsyncTaskFailed,
  isAsyncTaskDataAvailable,
  isAsyncTaskStarting,
} from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";

import { PAYMENT_OPTIONS_ADD_NEW } from "src/utils/constants";

export function CreatePaymentOption() {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();

  const navigate = useNavigate();
  const { message } = App.useApp();

  const [paymentConfigurations, setPaymentConfigurations] = useState<
    AsyncTask<PaymentConfigurations, AppError>
  >({
    status: "pending",
  });

  const fetchPaymentConfigurations = useCallback(
    async (signal?: AbortSignal) => {
      setPaymentConfigurations((previousConfigurations) =>
        isAsyncTaskDataAvailable(previousConfigurations)
          ? { data: previousConfigurations.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await getPaymentConfigurations({
        env,
        signal,
      });
      if (response.success) {
        setPaymentConfigurations({
          data: response.data,
          status: "successful",
        });
      } else {
        if (!isAbortedError(response.error)) {
          setPaymentConfigurations({ error: response.error, status: "failed" });
        }
      }
    },
    [env]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchPaymentConfigurations);

    return aborter;
  }, [fetchPaymentConfigurations]);

  const handleSubmit = (formValues: PaymentOptionFormData) => {
    const parsedFormData = paymentOptionFormParser.safeParse(formValues);

    if (parsedFormData.success) {
      return void createPaymentOption({
        env,
        identifier,
        payload: parsedFormData.data,
      }).then((response) => {
        if (response.success) {
          void message.success("Payment option added successfully");
          navigate(ROUTES.paymentOptions.path);
        } else {
          void message.error(response.error.message);
        }
      });
    } else {
      void notifyParseError(parsedFormData.error);
    }
  };

  return (
    <SiderLayoutContent
      description="Create a new payment option"
      showBackButton
      showDivider
      title={PAYMENT_OPTIONS_ADD_NEW}
    >
      <Card className="centered" title="Payment option details">
        {(() => {
          if (hasAsyncTaskFailed(paymentConfigurations)) {
            return (
              <Card className="centered">
                <ErrorResult
                  error={[
                    "An error occurred while downloading a payments configuration from the API:",
                    paymentConfigurations.error.message,
                  ].join("\n")}
                />
              </Card>
            );
          } else if (isAsyncTaskStarting(paymentConfigurations)) {
            return (
              <Card className="centered">
                <LoadingResult />
              </Card>
            );
          } else {
            return (
              <PaymentOptionForm
                initialValies={{
                  description: "",
                  name: "",
                  paymentOptions: [],
                }}
                onSubmit={handleSubmit}
                paymentConfigurations={paymentConfigurations.data}
              />
            );
          }
        })()}
      </Card>
    </SiderLayoutContent>
  );
}
