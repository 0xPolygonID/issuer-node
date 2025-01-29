import {
  Button,
  Card,
  Flex,
  Row,
  Space,
  Table,
  TableColumnsType,
  Tag,
  Tooltip,
  Typography,
} from "antd";
import { useCallback, useEffect, useState } from "react";
import { generatePath, useNavigate } from "react-router-dom";
import { getKeys } from "src/adapters/api/keys";
import { getPaymentConfigurations } from "src/adapters/api/payments";
import { notifyErrors } from "src/adapters/parsers";
import IconCheckMark from "src/assets/icons/check.svg?react";
import IconCopy from "src/assets/icons/copy-01.svg?react";
import EditIcon from "src/assets/icons/edit-02.svg?react";
import IconTrash from "src/assets/icons/trash-01.svg?react";
import { TableCard } from "src/components/shared/TableCard";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { AppError, Key, PaymentConfig, PaymentConfigurations } from "src/domain";
import { ROUTES } from "src/routes";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";

export function PaymentConfigTable({
  configs,
  onDelete,
  onEdit,
  showTitle,
}: {
  configs: PaymentConfig[];
  onDelete?: (index: number) => void;
  onEdit?: (index: number) => void;
  showTitle: boolean;
}) {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();
  const navigate = useNavigate();

  const [keys, setKeys] = useState<AsyncTask<Key[], AppError>>({
    status: "pending",
  });

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

  const fetchKeys = useCallback(
    async (signal?: AbortSignal) => {
      setKeys((previousKeys) =>
        isAsyncTaskDataAvailable(previousKeys)
          ? { data: previousKeys.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await getKeys({
        env,
        identifier,
        params: {},
        signal,
      });
      if (response.success) {
        setKeys({
          data: response.data.items.successful,
          status: "successful",
        });

        void notifyErrors(response.data.items.failed);
      } else {
        if (!isAbortedError(response.error)) {
          setKeys({ error: response.error, status: "failed" });
        }
      }
    },
    [env, identifier]
  );

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchKeys);

    return aborter;
  }, [fetchKeys]);

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchPaymentConfigurations);

    return aborter;
  }, [fetchPaymentConfigurations]);

  const tableColumns: TableColumnsType<PaymentConfig> = [
    {
      dataIndex: "paymentOptionID",
      key: "paymentOptionID",
      render: (paymentOptionID: PaymentConfig["paymentOptionID"]) => {
        const paymentOptionName =
          isAsyncTaskDataAvailable(paymentConfigurations) &&
          paymentConfigurations.data?.[`${paymentOptionID}`]?.PaymentOption.Name;
        return <Typography.Text>{paymentOptionName || paymentOptionID}</Typography.Text>;
      },
      title: "Name",
    },
    {
      dataIndex: "amount",
      key: "amount",
      render: (amount: PaymentConfig["amount"]) => <Typography.Text>{amount}</Typography.Text>,
      title: "Amount",
    },
    {
      dataIndex: "recipient",
      key: "recipient",
      render: (recipient: PaymentConfig["recipient"]) => (
        <Tooltip title={recipient}>
          <Typography.Text
            copyable={{
              icon: [<IconCopy key={0} />, <IconCheckMark key={1} />],
            }}
            ellipsis={{
              suffix: recipient.slice(-5),
            }}
          >
            {recipient}
          </Typography.Text>
        </Tooltip>
      ),
      title: "Recipient",
    },
    {
      dataIndex: "signingKeyID",
      key: "signingKeyID",
      render: (signingKeyID: PaymentConfig["signingKeyID"]) => {
        const keyName =
          isAsyncTaskDataAvailable(keys) && keys.data.find(({ id }) => id === signingKeyID)?.name;

        return (
          <Typography.Link
            onClick={() =>
              navigate(
                generatePath(ROUTES.keyDetails.path, {
                  keyID: signingKeyID,
                })
              )
            }
            strong
          >
            {keyName || signingKeyID}
          </Typography.Link>
        );
      },
      title: "Signin key",
    },
  ];

  if (onEdit || onDelete) {
    tableColumns.push({
      key: "actions",
      render: (_, __, index) => {
        return (
          <Flex>
            {onEdit && (
              <Button onClick={() => onEdit(index)} type="text">
                <EditIcon className="icon-secondary" />
              </Button>
            )}
            {onDelete && (
              <Button onClick={() => onDelete(index)} type="text">
                <IconTrash className="icon-secondary" />
              </Button>
            )}
          </Flex>
        );
      },
    });
  }

  return (
    <TableCard
      defaultContents={<Typography.Text strong>No configurations</Typography.Text>}
      isLoading={isAsyncTaskStarting(keys) || isAsyncTaskStarting(paymentConfigurations)}
      showDefaultContents={!configs.length}
      table={
        <Table
          columns={tableColumns}
          dataSource={configs}
          pagination={false}
          rowKey={(record) => record.paymentOptionID}
          showSorterTooltip
          sortDirections={["ascend", "descend"]}
          tableLayout="fixed"
        />
      }
      title={
        showTitle && (
          <Row justify="space-between">
            <Space size="middle">
              <Card.Meta title="Configurations" />

              <Tag>{configs.length}</Tag>
            </Space>
          </Row>
        )
      }
    />
  );
}
