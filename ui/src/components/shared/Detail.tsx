import { Button, Col, Flex, Grid, Row, Tag, TagProps, Typography, message } from "antd";

import copy from "copy-to-clipboard";
import { useRef } from "react";
import IconCheckMark from "src/assets/icons/check.svg?react";
import IconCopy from "src/assets/icons/copy-01.svg?react";
import IconDownload from "src/assets/icons/download-01.svg?react";
import { DownloadQRLink } from "src/components/shared/DownloadQRLink";

export function Detail({
  copyable,
  copyableText,
  donwloadLink,
  ellipsisPosition,
  href,
  label,
  tag,
  text,
}: {
  copyable?: boolean;
  copyableText?: string;
  donwloadLink?: boolean;
  ellipsisPosition?: number;
  href?: string;
  label: string;
  tag?: TagProps;
  text: string;
}) {
  const { sm } = Grid.useBreakpoint();
  const [messageAPI, messageContext] = message.useMessage();
  const donwloadLinkRef = useRef<HTMLButtonElement | null>(null);

  const onCopyToClipboard = (link: string) => {
    const hasCopied = copy(link, {
      format: "text/plain",
    });

    if (hasCopied) {
      void messageAPI.success("Link copied to clipboard.");
    } else {
      void messageAPI.error("Couldn't copy link. Please try again.");
    }
  };
  const value = ellipsisPosition ? text.slice(0, text.length - ellipsisPosition) : text;
  const element = (
    <Typography.Text
      copyable={
        copyable && {
          icon: [<IconCopy key={0} />, <IconCheckMark key={1} />],
          text: copyableText || text,
        }
      }
      ellipsis={ellipsisPosition ? { suffix: text.slice(-ellipsisPosition) } : true}
      style={{ textAlign: sm ? "right" : "left", width: "100%" }}
    >
      {tag ? (
        <Tag {...tag} style={{ marginInlineEnd: "initial" }}>
          {value}
        </Tag>
      ) : (
        value
      )}
    </Typography.Text>
  );

  return (
    <>
      {messageContext}
      <Row justify="space-between">
        <Col sm={10} xs={24}>
          <Typography.Text type="secondary">{label}</Typography.Text>
        </Col>
        <Col sm={14} xs={24}>
          {(() => {
            if (donwloadLink && href) {
              return (
                <Flex align="center" gap={8} justify="flex-end">
                  <DownloadQRLink
                    button={
                      <Button
                        icon={<IconDownload />}
                        iconPosition="end"
                        ref={donwloadLinkRef}
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
            } else if (href) {
              return (
                <Typography.Link ellipsis href={href} target="_blank">
                  {element}
                </Typography.Link>
              );
            } else {
              return element;
            }
          })()}
        </Col>
      </Row>
    </>
  );
}
