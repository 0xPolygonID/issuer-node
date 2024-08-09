import { Typography } from "antd";
import { useIssuerContext } from "src/contexts/Issuer";
import { IssuerIdentifier } from "src/domain";

function formatIdentifier(identifier: IssuerIdentifier): string {
  if (identifier) {
    const parts = identifier.split(":");
    const id = parts.at(-1);
    const shortId = `${id?.slice(0, 5)}...${id?.slice(-4)}`;
    return parts.toSpliced(-1, 1, shortId).join(":");
  }

  return "Select issuer";
}

export function SelectedIssuer() {
  const { issuerIdentifier } = useIssuerContext();

  return (
    <Typography.Text style={{ fontSize: 12, whiteSpace: "nowrap" }}>
      {formatIdentifier(issuerIdentifier)}
    </Typography.Text>
  );
}
