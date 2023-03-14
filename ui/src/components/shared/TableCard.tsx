import { Card, Row, Space } from "antd";
import { ReactNode } from "react";

import { LoadingResult } from "src/components/shared/LoadingResult";
import { SearchBox } from "src/components/shared/SearchBox";

export function TableCard({
  defaultContents,
  isLoading,
  onSearch,
  query,
  searchPlaceholder,
  showDefaultContents,
  table,
  title,
}: {
  defaultContents: ReactNode;
  isLoading: boolean;
  onSearch: (value: string) => void;
  query: string | null;
  searchPlaceholder: string;
  showDefaultContents: boolean;
  table: ReactNode;
  title: ReactNode;
}) {
  return (
    <Card bodyStyle={{ padding: 0 }} title={title}>
      {!showDefaultContents && (
        <Row style={{ padding: "16px 12px", width: "60%" }}>
          <SearchBox onSearch={onSearch} placeholder={searchPlaceholder} query={query} />
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
