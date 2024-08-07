import { Typography } from "antd";
import { useIssuerContext } from "src/contexts/Issuer";
import { Identifier } from "src/domain/identifier";

function formatIdentifier(identifier: Identifier): string {
  if (identifier) {
    const parts = identifier.split(":");
    const id = parts.at(-1);
    const shortId = `${id?.slice(0, 5)}...${id?.slice(-4)}`;
    return parts.toSpliced(-1, 1, shortId).join(":");
  }

  return "Select issuer";
}

export function SelectedIssuer() {
  const { identifier } = useIssuerContext();

  return (
    <Typography.Text style={{ fontSize: 12, whiteSpace: "nowrap" }}>
      {formatIdentifier(identifier)}
    </Typography.Text>
  );
}
