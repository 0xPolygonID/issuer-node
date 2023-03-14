import { Button, Card, Grid, Image, Space, Typography } from "antd";

export function Explainer({
  CTA,
  description,
  handleDismiss,
  image,
  title,
}: {
  CTA: { label: string; url: string };
  description: string;
  handleDismiss: () => void;
  image: string;
  title: string;
}) {
  const { xl } = Grid.useBreakpoint();
  const { label, url } = CTA;

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

          <Button onClick={handleDismiss} type="primary">
            Dismiss
          </Button>
        </Space>
      </Space>
    </Card>
  );
}
