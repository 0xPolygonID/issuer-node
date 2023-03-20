import { Button, Card, Grid, Image, Space, Typography } from "antd";
import { useState } from "react";
import { z } from "zod";

import { getStorageByKey, setStorageByKey } from "src/utils/localStorage";

export function Explainer({
  CTA,
  description,
  image,
  isDismissable,
  localStorageKey,
  title,
}: {
  CTA: { label: string; url: string };
  description: string;
  image: string;
  isDismissable: boolean;
  localStorageKey: string;
  title: string;
}) {
  const [isShowing, setShowing] = useState(
    !isDismissable
      ? true
      : getStorageByKey({ defaultValue: true, key: localStorageKey, parser: z.boolean() })
  );

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

          {isDismissable && (
            <Button
              onClick={() => {
                setShowing(false);
                setStorageByKey({ key: localStorageKey, value: false });
              }}
              type="primary"
            >
              Dismiss
            </Button>
          )}
        </Space>
      </Space>
    </Card>
  );
}
