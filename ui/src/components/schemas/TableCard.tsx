import { Card, Row, Space } from "antd";
import { ReactNode } from "react";

import { SearchBox } from "src/components/schemas/SearchBox";
import { LoadingResult } from "src/components/shared/LoadingResult";

export function TableCard({
  defaultContents,
  isLoading,
  onSearch,
  query,
  showDefaultContents,
  table,
  title,
}: {
  defaultContents: ReactNode;
  isLoading: boolean;
  onSearch: (value: string) => void;
  query: string | null;
  showDefaultContents: boolean;
  table: ReactNode;
  title: ReactNode;
}) {
  return (
    <Card bodyStyle={{ padding: 0 }} title={title}>
      {!showDefaultContents && (
        <Row style={{ padding: "16px 12px", width: "60%" }}>
          <SearchBox onSearch={onSearch} query={query} />
        </Row>
      )}

      {isLoading ? (
        <LoadingResult />
      ) : showDefaultContents ? (
        <Space align="center" direction="vertical" size="middle" style={{ padding: 24 }}>
          {defaultContents}
        </Space>
      ) : (
        table
      )}
    </Card>
  );
}
