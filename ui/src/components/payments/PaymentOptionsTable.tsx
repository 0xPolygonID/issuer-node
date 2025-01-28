import {
  App,
  Avatar,
  Button,
  Card,
  Dropdown,
  Row,
  Space,
  Table,
  TableColumnsType,
  Tag,
  Tooltip,
  Typography,
} from "antd";
import { ItemType } from "antd/es/menu/interface";
import { useCallback, useEffect, useState } from "react";
import { Link, generatePath, useNavigate, useSearchParams } from "react-router-dom";

import { deletePaymentOption, getPaymentOptions } from "src/adapters/api/payments";
import { notifyErrors, positiveIntegerFromStringParser } from "src/adapters/parsers";
import IconDots from "src/assets/icons/dots-vertical.svg?react";
import IconInfoCircle from "src/assets/icons/info-circle.svg?react";
import IconPaymentOptions from "src/assets/icons/payment-options.svg?react";
import IconPlus from "src/assets/icons/plus.svg?react";
import { DeleteItem } from "src/components/schemas/DeleteItem";
import { ErrorResult } from "src/components/shared/ErrorResult";
import { NoResults } from "src/components/shared/NoResults";
import { TableCard } from "src/components/shared/TableCard";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { AppError, PaymentOption } from "src/domain";
import { ROUTES } from "src/routes";
import { AsyncTask, isAsyncTaskDataAvailable, isAsyncTaskStarting } from "src/utils/async";
import { isAbortedError, makeRequestAbortable } from "src/utils/browser";
import {
  DEFAULT_PAGINATION_MAX_RESULTS,
  DEFAULT_PAGINATION_PAGE,
  DEFAULT_PAGINATION_TOTAL,
  DETAILS,
  DOTS_DROPDOWN_WIDTH,
  PAGINATION_MAX_RESULTS_PARAM,
  PAGINATION_PAGE_PARAM,
  PAYMENT_OPTIONS_ADD_NEW,
  QUERY_SEARCH_PARAM,
} from "src/utils/constants";
import { formatDate } from "src/utils/forms";

export function PaymentOptionsTable() {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();
  const { message } = App.useApp();
  const navigate = useNavigate();

  const [paymentOptions, setPaymentOptions] = useState<AsyncTask<PaymentOption[], AppError>>({
    status: "pending",
  });

  const [searchParams, setSearchParams] = useSearchParams();
  const queryParam = searchParams.get(QUERY_SEARCH_PARAM);
  const paginationPageParam = searchParams.get(PAGINATION_PAGE_PARAM);
  const paginationMaxResultsParam = searchParams.get(PAGINATION_MAX_RESULTS_PARAM);

  const paginationPageParsed = positiveIntegerFromStringParser.safeParse(paginationPageParam);
  const paginationMaxResultsParsed =
    positiveIntegerFromStringParser.safeParse(paginationMaxResultsParam);

  const [paginationTotal, setPaginationTotal] = useState<number>(DEFAULT_PAGINATION_TOTAL);
  const paginationPage = paginationPageParsed.success
    ? paginationPageParsed.data
    : DEFAULT_PAGINATION_PAGE;
  const paginationMaxResults = paginationMaxResultsParsed.success
    ? paginationMaxResultsParsed.data
    : DEFAULT_PAGINATION_MAX_RESULTS;

  const paymentOptionsList = isAsyncTaskDataAvailable(paymentOptions) ? paymentOptions.data : [];
  const showDefaultContent =
    paymentOptions.status === "successful" && paymentOptionsList.length === 0;

  const tableColumns: TableColumnsType<PaymentOption> = [
    {
      dataIndex: "name",
      key: "name",
      render: (name: PaymentOption["name"], { description, id }: PaymentOption) => (
        <Typography.Link
          onClick={() =>
            navigate(
              generatePath(ROUTES.paymentOptionDetails.path, {
                paymentOptionID: id,
              })
            )
          }
          strong
        >
          <Tooltip title={description}>{name}</Tooltip>
        </Typography.Link>
      ),
      title: "Name",
    },
    {
      dataIndex: "createdAt",
      key: "createdAt",
      render: (createdAt: PaymentOption["createdAt"]) => (
        <Typography.Text>{formatDate(createdAt)}</Typography.Text>
      ),
      sorter: ({ createdAt: a }, { createdAt: b }) => b.getTime() - a.getTime(),
      title: "Created date",
    },
    {
      dataIndex: "modifiedAt",
      key: "modifiedAt",
      render: (modifiedAt: PaymentOption["modifiedAt"]) => (
        <Typography.Text>{formatDate(modifiedAt)}</Typography.Text>
      ),
      sorter: ({ modifiedAt: a }, { modifiedAt: b }) => b.getTime() - a.getTime(),
      title: "Modified date",
    },

    {
      dataIndex: "id",
      key: "id",
      render: (id: PaymentOption["id"]) => {
        const items: Array<ItemType> = [
          {
            icon: <IconInfoCircle />,
            key: "details",
            label: DETAILS,
            onClick: () =>
              navigate(
                generatePath(ROUTES.paymentOptionDetails.path, {
                  paymentOptionID: id,
                })
              ),
          },
          {
            key: "divider1",
            type: "divider",
          },
          {
            danger: true,
            key: "delete",
            label: (
              <DeleteItem
                onOk={() => handleDeletePaymentOption(id)}
                title="Are you sure you want to delete this payment option?"
              />
            ),
          },
        ];

        return (
          <Dropdown
            menu={{
              items,
            }}
          >
            <Row>
              <IconDots className="icon-secondary" />
            </Row>
          </Dropdown>
        );
      },

      width: DOTS_DROPDOWN_WIDTH,
    },
  ];

  const updateUrlParams = useCallback(
    ({ maxResults, page }: { maxResults?: number; page?: number }) => {
      setSearchParams((previousParams) => {
        const params = new URLSearchParams(previousParams);
        params.set(
          PAGINATION_PAGE_PARAM,
          page !== undefined ? page.toString() : DEFAULT_PAGINATION_PAGE.toString()
        );
        params.set(
          PAGINATION_MAX_RESULTS_PARAM,
          maxResults !== undefined
            ? maxResults.toString()
            : DEFAULT_PAGINATION_MAX_RESULTS.toString()
        );

        return params;
      });
    },
    [setSearchParams]
  );

  const fetchPaymentOptions = useCallback(
    async (signal?: AbortSignal) => {
      setPaymentOptions((previousPaymentOptions) =>
        isAsyncTaskDataAvailable(previousPaymentOptions)
          ? { data: previousPaymentOptions.data, status: "reloading" }
          : { status: "loading" }
      );

      const response = await getPaymentOptions({
        env,
        identifier,
        params: {
          maxResults: paginationMaxResults,
          page: paginationPage,
        },
        signal,
      });
      if (response.success) {
        setPaymentOptions({
          data: response.data.items.successful,
          status: "successful",
        });
        setPaginationTotal(response.data.meta.total);
        updateUrlParams({
          maxResults: response.data.meta.max_results,
          page: response.data.meta.page,
        });
        void notifyErrors(response.data.items.failed);
      } else {
        if (!isAbortedError(response.error)) {
          setPaymentOptions({ error: response.error, status: "failed" });
        }
      }
    },
    [env, paginationMaxResults, paginationPage, identifier, updateUrlParams]
  );

  const handleDeletePaymentOption = (paymentOptionID: string) => {
    void deletePaymentOption({ env, identifier, paymentOptionID }).then((response) => {
      if (response.success) {
        void fetchPaymentOptions();
        void message.success(response.data.message);
      } else {
        void message.error(response.error.message);
      }
    });
  };

  useEffect(() => {
    const { aborter } = makeRequestAbortable(fetchPaymentOptions);

    return aborter;
  }, [fetchPaymentOptions]);

  return (
    <TableCard
      defaultContents={
        <>
          <Avatar className="avatar-color-icon" icon={<IconPaymentOptions />} size={48} />

          <Typography.Text strong>No payment options</Typography.Text>

          <Typography.Text type="secondary">
            Your payment options will be listed here.
          </Typography.Text>

          <Link to={ROUTES.createPaymentOption.path}>
            <Button icon={<IconPlus />} type="primary">
              {PAYMENT_OPTIONS_ADD_NEW}
            </Button>
          </Link>
        </>
      }
      isLoading={isAsyncTaskStarting(paymentOptions)}
      query={queryParam}
      showDefaultContents={showDefaultContent}
      table={
        <Table
          columns={tableColumns}
          dataSource={paymentOptionsList}
          locale={{
            emptyText:
              paymentOptions.status === "failed" ? (
                <ErrorResult error={paymentOptions.error.message} />
              ) : (
                <NoResults searchQuery={queryParam} />
              ),
          }}
          onChange={({ current, pageSize, total }) => {
            setPaginationTotal(total || DEFAULT_PAGINATION_TOTAL);
            updateUrlParams({
              maxResults: pageSize,
              page: current,
            });
          }}
          pagination={{
            current: paginationPage,
            hideOnSinglePage: true,
            pageSize: paginationMaxResults,
            position: ["bottomRight"],
            total: paginationTotal,
          }}
          rowKey="id"
          showSorterTooltip
          sortDirections={["ascend", "descend"]}
        />
      }
      title={
        <Row justify="space-between">
          <Space size="middle">
            <Card.Meta title="Payment options" />
            <Tag>{paginationTotal}</Tag>
          </Space>
        </Row>
      }
    />
  );
}
