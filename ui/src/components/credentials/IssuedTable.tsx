import { Avatar, Button, Card, Row, Space, Table, Tag, Typography } from "antd";
import { Link, generatePath } from "react-router-dom";

import { ReactComponent as IconCreditCardPlus } from "src/assets/icons/credit-card-plus.svg";
import { ReactComponent as IconCreditCardRefresh } from "src/assets/icons/credit-card-refresh.svg";

import { TableCard } from "src/components/shared/TableCard";
import { ROUTES } from "src/routes";
import { ISSUED, ISSUE_CREDENTIAL } from "src/utils/constants";

export function IssuedTable() {
  return (
    <TableCard
      defaultContents={
        <>
          <Avatar className="avatar-color-cyan" icon={<IconCreditCardRefresh />} size={48} />

          <Typography.Text strong>No credentials</Typography.Text>
          <Typography.Text type="secondary">
            Issued credentials will be listed here.
          </Typography.Text>

          <Link to={generatePath(ROUTES.issueCredential.path)}>
            <Button icon={<IconCreditCardPlus />} type="primary">
              {ISSUE_CREDENTIAL}
            </Button>
          </Link>
        </>
      }
      isLoading={false}
      onSearch={() => null}
      query={null}
      searchPlaceholder="Search credentials, attributes, identifiers..."
      showDefaultContents={true}
      table={
        <Table
          dataSource={[]}
          pagination={false}
          rowKey="id"
          showSorterTooltip
          sortDirections={["ascend", "descend"]}
        />
      }
      title={
        <Row justify="space-between">
          <Space size="middle">
            <Card.Meta title={ISSUED} />

            <Tag color="blue">{0}</Tag>
          </Space>
        </Row>
      }
    />
  );
}
