import { App, Button, Card, Dropdown, Flex, Row, Space, Typography } from "antd";
import { useCallback, useEffect, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";

import { useIdentityContext } from "../../contexts/Identity";
import {
  deletePaymentOption,
  getPaymentConfigurations,
  getPaymentOption,
  updatePaymentOption,
} from "src/adapters/api/payments";
import { buildAppError, notifyError, notifyParseError } from "src/adapters/parsers";
import { PaymentOptionFormData, paymentOptionFormParser } from "src/adapters/parsers/view";
import IconDots from "src/assets/icons/dots-vertical.svg?react";
import EditIcon from "src/assets/icons/edit-02.svg?react";
import { PaymentConfigTable } from "src/components/payments/PaymentConfigTable";
import { PaymentOptionForm } from "src/components/payments/PaymentOptionForm";
import { DeleteItem } from "src/components/schemas/DeleteItem";
import { Detail } from "src/components/shared/Detail";
import { EditModal } from "src/components/shared/EditModal";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { LoadingResult } from "src/components/shared/LoadingResult";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { AppError, PaymentConfigurations, PaymentOption as PaymentOptionType } from "src/domain";
import { ROUTES } from "src/routes";
import {
  AsyncTask,
  hasAsyncTaskFailed,
  isAsyncTaskDataAvailable,
  isAsyncTaskStarting,
} from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import { PAYMENT_OPTIONS_DETAILS } from "src/utils/constants";
import { formatDate } from "src/utils/forms";

export function PaymentOption() {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();
  const { message } = App.useApp();
  const navigate = useNavigate();

  const [isEditModalOpen, setIsEditModalOpen] = useState(false);
  const [paymentOption, setPaymentOption] = useState<AsyncTask<PaymentOptionType, AppError>>({
    status: "pending",
  });

  const [paymentConfigurations, setPaymentConfigurations] = useState<
    AsyncTask<PaymentConfigurations, AppError>
  >({
    status: "pending",
  });

  const { paymentOptionID } = useParams();

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

  const fetchPaymentOption = useCallback(
    async (signal?: AbortSignal) => {
      if (paymentOptionID) {
        setPaymentOption({ status: "loading" });

        const response = await getPaymentOption({
          env,
          identifier,
          paymentOptionID,
          signal,
        });

        if (response.success) {
          setPaymentOption({ data: response.data, status: "successful" });
          void fetchPaymentConfigurations(signal);
        } else {
          if (!isAbortedError(response.error)) {
            setPaymentOption({ error: response.error, status: "failed" });
          }
        }
      }
    },
    [env, paymentOptionID, identifier, fetchPaymentConfigurations]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchPaymentOption);

    return aborter;
  }, [fetchPaymentOption]);

  if (!paymentOptionID) {
    return <ErrorResult error="No payment option provided." />;
  }

  const handleDeletePaymentOption = () => {
    void deletePaymentOption({ env, identifier, paymentOptionID }).then((response) => {
      if (response.success) {
        navigate(ROUTES.paymentOptions.path);
        void message.success(response.data.message);
      } else {
        void message.error(response.error.message);
      }
    });
  };

  const handleEdit = (formValues: PaymentOptionFormData) => {
    const parsedFormData = paymentOptionFormParser.safeParse(formValues);

    if (parsedFormData.success) {
      return void updatePaymentOption({
        env,
        identifier,
        payload: parsedFormData.data,
        paymentOptionID,
      }).then((response) => {
        if (response.success) {
          void fetchPaymentOption().then(() => {
            setIsEditModalOpen(false);
            void message.success("Payment option edited successfully");
          });
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
      description="View payment option details"
      showBackButton
      showDivider
      title={PAYMENT_OPTIONS_DETAILS}
    >
      {(() => {
        if (hasAsyncTaskFailed(paymentOption) || hasAsyncTaskFailed(paymentConfigurations)) {
          return (
            <Card className="centered">
              {hasAsyncTaskFailed(paymentOption) && (
                <ErrorResult
                  error={[
                    "An error occurred while downloading a payment option from the API:",
                    paymentOption.error.message,
                  ].join("\n")}
                />
              )}
              {hasAsyncTaskFailed(paymentConfigurations) && (
                <ErrorResult
                  error={[
                    "An error occurred while downloading a payments configuration from the API:",
                    paymentConfigurations.error.message,
                  ].join("\n")}
                />
              )}
            </Card>
          );
        } else if (
          isAsyncTaskStarting(paymentOption) ||
          isAsyncTaskStarting(paymentConfigurations)
        ) {
          return (
            <Card className="centered">
              <LoadingResult />
            </Card>
          );
        } else {
          return (
            <>
              <Card
                className="centered"
                title={
                  <Flex align="center" gap={8} justify="space-between">
                    <Typography.Text style={{ fontWeight: 600 }}>
                      {paymentOption.data.name}
                    </Typography.Text>
                    <Flex gap={8}>
                      <Button
                        icon={<EditIcon />}
                        onClick={() => setIsEditModalOpen(true)}
                        style={{ flexShrink: 0 }}
                        type="text"
                      />

                      <Dropdown
                        menu={{
                          items: [
                            {
                              danger: true,
                              key: "delete",
                              label: (
                                <DeleteItem
                                  onOk={handleDeletePaymentOption}
                                  title="Are you sure you want to delete this payment option?"
                                />
                              ),
                            },
                          ],
                        }}
                      >
                        <Row>
                          <IconDots className="icon-secondary" />
                        </Row>
                      </Dropdown>
                    </Flex>
                  </Flex>
                }
              >
                <Flex gap={24} vertical>
                  <Card className="background-grey">
                    <Space direction="vertical">
                      <Detail label="Name" text={paymentOption.data.name} />
                      <Detail label="Description" text={paymentOption.data.description} />
                      <Detail
                        label="Created date"
                        text={formatDate(paymentOption.data.createdAt)}
                      />
                      <Detail
                        label="Modified date"
                        text={formatDate(paymentOption.data.modifiedAt)}
                      />
                    </Space>
                  </Card>

                  <PaymentConfigTable
                    configs={paymentOption.data.paymentOptions.map(
                      ({ amount, paymentOptionID, ...other }) => {
                        const configuration = paymentConfigurations.data[paymentOptionID];
                        if (!configuration) {
                          void notifyError(
                            buildAppError(
                              `Can't find payment configuration for ID: ${paymentOptionID}`
                            )
                          );
                        }

                        return {
                          amount: configuration
                            ? (
                                parseFloat(amount) /
                                Math.pow(10, configuration.PaymentOption.Decimals)
                              ).toString()
                            : amount,
                          paymentOptionID,
                          ...other,
                        };
                      }
                    )}
                    showTitle={true}
                  />
                </Flex>
              </Card>

              <EditModal
                onClose={() => setIsEditModalOpen(false)}
                open={isEditModalOpen}
                title="Edit payment option"
              >
                <PaymentOptionForm
                  initialValies={{
                    description: paymentOption.data.description,
                    name: paymentOption.data.name,
                    paymentOptions: paymentOption.data.paymentOptions
                      .map(({ amount, paymentOptionID, ...other }) => {
                        const configuration = paymentConfigurations.data[paymentOptionID];
                        return configuration
                          ? {
                              amount: (
                                parseFloat(amount) /
                                Math.pow(10, configuration.PaymentOption.Decimals)
                              ).toString(),
                              decimals: configuration.PaymentOption.Decimals,
                              paymentOptionID: paymentOptionID.toString(),
                              ...other,
                            }
                          : null;
                      })
                      .filter((option) => !!option),
                  }}
                  onSubmit={handleEdit}
                  paymentConfigurations={paymentConfigurations.data}
                />
              </EditModal>
            </>
          );
        }
      })()}
    </SiderLayoutContent>
  );
}
