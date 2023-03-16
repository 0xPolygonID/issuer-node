import { Button, Card, Grid, Image, Space, Typography } from "antd";

import { useLocalStorage } from "src/hooks/useLocalStorage";
import { kebabCase } from "src/utils/string";

export function Explainer({
  CTA,
  description,
  image,
  title,
}: {
  CTA: { label: string; url: string };
  description: string;
  image: string;
  title: string;
}) {
  const [isShowing, setShowing] = useLocalStorage(`explainer-${kebabCase(title)}`, true);

  const { xl } = Grid.useBreakpoint();
  const { label, url } = CTA;

  if (!isShowing) {
    return null;
  }

  return (
    <Card className="explainer" title={title}>
      {xl && <Image preview={false} src={image} />}

      <Space direction="vertical" size="large" style={{ position: "relative" }}>
        <Typography>{description}</Typography>

        <Space>
          {label && url && (
            <Button href={url} target="_blank">
              {label}
            </Button>
          )}

          <Button onClick={() => setShowing(false)} type="primary">
            Dismiss
          </Button>
        </Space>
      </Space>
    </Card>
  );
}
