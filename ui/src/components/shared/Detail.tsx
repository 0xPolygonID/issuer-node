import { App, Button, Col, Flex, Grid, Row, Tag, TagProps, Typography } from "antd";

import copy from "copy-to-clipboard";
import { useRef } from "react";
import IconCheckMark from "src/assets/icons/check.svg?react";
import IconCopy from "src/assets/icons/copy-01.svg?react";
import IconDownload from "src/assets/icons/download-01.svg?react";
import { DownloadQRLink } from "src/components/shared/DownloadQRLink";

export function Detail({
  copyable,
  copyableText,
  downloadLink,
  ellipsisPosition,
  href,
  label,
  tag,
  text,
}: {
  copyable?: boolean;
  copyableText?: string;
  downloadLink?: boolean;
  ellipsisPosition?: number;
  href?: string;
  label: string;
  tag?: TagProps;
  text: string;
}) {
  const { sm } = Grid.useBreakpoint();
  const { message } = App.useApp();
  const downloadLinkRef = useRef<HTMLButtonElement | null>(null);

  const onCopyToClipboard = (link: string) => {
    const hasCopied = copy(link, {
      format: "text/plain",
    });

    if (hasCopied) {
      void message.success("Link copied to clipboard.");
    } else {
      void message.error("Couldn't copy link. Please try again.");
    }
  };
  const value = ellipsisPosition ? text.slice(0, text.length - ellipsisPosition) : text;
  const Component: React.ElementType = href ? Typography.Link : Typography.Text;

  const componentProps =
    Component === Typography.Link
      ? {
          ellipsis: true,
          href,
          target: "_blank",
        }
      : {
          ellipsis: ellipsisPosition ? { suffix: text.slice(-ellipsisPosition) } : true,
        };

  const element = (
    <Flex
      justify={sm ? "flex-end" : "flex-start"}
      style={{ display: "inline-flex", width: "100%" }}
    >
      <Component
        {...componentProps}
        style={{
          textAlign: sm ? "right" : "left",
        }}
      >
        {tag ? (
          <Tag {...tag} style={{ marginInlineEnd: "initial" }}>
            {value}
          </Tag>
        ) : (
          value
        )}
      </Component>
      {copyable && (
        <Typography.Text
          copyable={
            copyable && {
              icon: [<IconCopy key={0} />, <IconCheckMark key={1} />],
              text: copyableText || text,
            }
          }
        />
      )}
    </Flex>
  );

  return (
    <Row justify="space-between">
      <Col sm={10} xs={24}>
        <Typography.Text type="secondary">{label}</Typography.Text>
      </Col>
      <Col sm={14} xs={24}>
        {(() => {
          if (downloadLink && href) {
            return (
              <Flex align="center" gap={8} justify={sm ? "flex-end" : "flex-start"}>
                <DownloadQRLink
                  button={
                    <Button
                      icon={<IconDownload />}
                      iconPosition="end"
                      ref={downloadLinkRef}
                      style={{ height: "auto", padding: 0 }}
                      type="link"
                    >
                      Download QR
                    </Button>
                  }
                  fileName={label}
                  hidden
                  link={href}
                />
                {copyable && (
                  <Button
                    icon={<IconCopy />}
                    iconPosition="end"
                    onClick={() => onCopyToClipboard(href)}
                    style={{ height: "auto", padding: 0 }}
                    type="link"
                  >
                    Copy link
                  </Button>
                )}
              </Flex>
            );
          } else {
            return element;
          }
        })()}
      </Col>
    </Row>
  );
}
