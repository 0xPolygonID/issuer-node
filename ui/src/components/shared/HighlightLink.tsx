import { App, Button, Card, Flex, Typography, theme } from "antd";
import copy from "copy-to-clipboard";
import IconCopy from "src/assets/icons/copy-01.svg?react";
import IconLink from "src/assets/icons/link-external-01.svg?react";

export function HighlightLink({ link, openable }: { link: string; openable: boolean }) {
  const { token } = theme.useToken();
  const { message } = App.useApp();

  const onCopyToClipboard = () => {
    const hasCopied = copy(link, {
      format: "text/plain",
    });

    if (hasCopied) {
      void message.success("Link copied to clipboard.");
    } else {
      void message.error("Couldn't copy link. Please try again.");
    }
  };

  return (
    <Flex gap={6} vertical>
      <Card bordered={false} className="background-grey" style={{ boxShadow: "none" }}>
        <Flex gap={8} style={{ float: "right", paddingLeft: 12 }}>
          {openable && (
            <Button
              href={link}
              icon={<IconLink />}
              style={{ borderColor: token.colorTextSecondary, color: token.colorTextSecondary }}
              target="_blank"
            />
          )}
          <Button
            icon={<IconCopy />}
            onClick={onCopyToClipboard}
            style={{ borderColor: token.colorTextSecondary, color: token.colorTextSecondary }}
          />
        </Flex>
        <Typography.Text
          style={{
            color: token.colorTextSecondary,
            fontFamily: "RobotoMono-Regular",
            fontSize: 12,
          }}
        >
          {link}
        </Typography.Text>
      </Card>
    </Flex>
  );
}
