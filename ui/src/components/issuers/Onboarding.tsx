import { Avatar, Card, Divider, Flex, Grid, Typography, message } from "antd";
import React from "react";
import { useNavigate } from "react-router-dom";
import { createIssuer } from "../../adapters/api/issuers";
import { IssuerFormData } from "src/adapters/parsers/view";
import IconCheck from "src/assets/icons/check.svg?react";
import IconIssue from "src/assets/icons/credential-card.svg?react";
import IconSchema from "src/assets/icons/file-search-02.svg?react";
import IconIdentity from "src/assets/icons/fingerprint-02.svg?react";

import { IssuerForm } from "src/components/issuers/IssuerForm";
import { useEnvContext } from "src/contexts/Env";

import { useIssuerContext } from "src/contexts/Issuer";
import { ROUTES } from "src/routes";

import { FINALISE_SETUP } from "src/utils/constants";

const cards = [
  {
    icon: <IconIssue />,
    text: "Issue verifiable credentials directly or via links",
    title: "Issue credentials",
  },
  {
    icon: <IconIdentity />,
    text: "Add new identities with different DIDs and settings",
    title: "Manage identities",
  },
  {
    icon: <IconSchema />,
    text: "Import custom schemas and use them to issue verifiable credentials",
    title: "Import custom schemas",
  },
];

export function Onboarding() {
  const env = useEnvContext();
  const { handleChange } = useIssuerContext();
  const navigate = useNavigate();
  const [messageAPI, messageContext] = message.useMessage();

  const { lg } = Grid.useBreakpoint();

  const handleSubmit = (formValues: IssuerFormData) =>
    void createIssuer({ env, payload: formValues }).then((response) => {
      if (response.success) {
        const {
          data: { identifier },
        } = response;
        void messageAPI.success("Identity added successfully");
        handleChange(identifier);
        navigate(ROUTES.schemas.path);
      } else {
        void messageAPI.error(response.error.message);
      }
    });

  return (
    <>
      {messageContext}

      <Flex className="onboarding" gap={32} style={{ padding: "0 24px" }} vertical>
        <Flex align="center" gap={8} style={{ textAlign: "center" }} vertical>
          <Avatar
            className="onboarding-check-icon"
            icon={<IconCheck />}
            size={48}
            style={{ marginBottom: 16 }}
          />

          <Typography.Text style={{ fontSize: 30 }}>
            You successfully installed Issuer Node
          </Typography.Text>
          <Typography.Text style={{ fontSize: 20 }}>
            Here&apos;s what you&apos;re going to be able to do with the issuer node, once you
            finalise your setup
          </Typography.Text>
        </Flex>

        <Flex gap={12} vertical={!lg}>
          {cards.map(({ icon, text, title }, index) => {
            const isLastCard = index + 1 === cards.length;
            return (
              <React.Fragment key={title}>
                <Card
                  style={{
                    boxShadow:
                      "0px 1px 3px 0px rgba(19, 19, 19, 0.1), 0px 1px 2px 0px rgba(19, 19, 19, 0.06)",
                    flex: 1,
                  }}
                >
                  <Flex gap={16}>
                    <Avatar
                      className="avatar-color-icon"
                      icon={icon}
                      size={48}
                      style={{ flexShrink: 0 }}
                    />

                    <Flex gap={4} vertical>
                      <Typography.Text strong style={{ fontSize: 18 }}>
                        {title}
                      </Typography.Text>
                      <Typography.Text style={{ fontSize: 16 }} type="secondary">
                        {text}
                      </Typography.Text>
                    </Flex>
                  </Flex>
                </Card>
                {!isLastCard && lg && <div className="dotted-divider" />}
              </React.Fragment>
            );
          })}
        </Flex>

        <Divider
          className="onboarding-divider"
          style={{ alignSelf: "center", borderWidth: 2, height: 48 }}
          type="vertical"
        />

        <Flex
          align="center"
          gap={24}
          style={{ alignSelf: "center", maxWidth: 700, paddingBottom: 24 }}
          vertical
        >
          <Typography.Text style={{ fontSize: 20 }}>
            Finalise the setup by adding a new identity
          </Typography.Text>

          <IssuerForm onSubmit={handleSubmit} submitBtnText={FINALISE_SETUP} />
        </Flex>
      </Flex>
    </>
  );
}
