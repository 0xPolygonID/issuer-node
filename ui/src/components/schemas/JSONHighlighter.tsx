import { Card } from "antd";
import SyntaxHighlighter from "react-syntax-highlighter";
import { a11yLight } from "react-syntax-highlighter/dist/esm/styles/hljs";
import { Json } from "src/domain";

export function JSONHighlighter({ json }: { json: Json }) {
  return (
    <Card className="background-grey">
      <SyntaxHighlighter
        customStyle={{
          background: "none",
          margin: 0,
        }}
        language="json"
        style={a11yLight}
      >
        {JSON.stringify(json, null, 2)}
      </SyntaxHighlighter>
    </Card>
  );
}
