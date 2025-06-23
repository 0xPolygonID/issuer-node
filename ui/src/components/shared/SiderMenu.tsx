import { App, Col, Menu, Row, Space, Tag, Typography } from "antd";
import { useState } from "react";
import { generatePath, matchRoutes, useLocation, useNavigate } from "react-router-dom";

import IconCredentials from "src/assets/icons/credit-card-refresh.svg?react";
import IconDisplayMethod from "src/assets/icons/display-method.svg?react";
import IconFile from "src/assets/icons/file-05.svg?react";
import IconSchema from "src/assets/icons/file-search-02.svg?react";
import IconIdentities from "src/assets/icons/fingerprint-02.svg?react";
import IconLink from "src/assets/icons/link-external-01.svg?react";
import IconPaymentOptions from "src/assets/icons/payment-options.svg?react";
import IconSettings from "src/assets/icons/settings-01.svg?react";
import IconIssuerState from "src/assets/icons/switch-horizontal.svg?react";
import IconConnections from "src/assets/icons/users-01.svg?react";
import { LogoLink } from "src/components/shared/LogoLink";
import { SettingsModal } from "src/components/shared/SettingsModal";
import { UserDisplay } from "src/components/shared/UserDisplay";
import { useEnvContext } from "src/contexts/Env";
import { useIssuerStateContext } from "src/contexts/IssuerState";
import { ROUTES } from "src/routes";
import { isAsyncTaskDataAvailable } from "src/utils/async";
import {
  CONNECTIONS,
  CREDENTIALS,
  CREDENTIALS_TABS,
  DISPLAY_METHODS,
  DOCS_URL,
  IDENTITIES,
  ISSUER_STATE,
  PAYMENT_OPTIONS,
  PAYMENT_REQUESTS,
  SCHEMAS,
} from "src/utils/constants";

export function SiderMenu({
  isBreakpoint,
  onClick,
}: {
  isBreakpoint?: boolean;
  onClick: () => void;
}) {
  const { buildTag, paymentPagesEnabled } = useEnvContext();
  const { status } = useIssuerStateContext();

  const { pathname } = useLocation();
  const navigate = useNavigate();
  const { message } = App.useApp();
  const [isSettingsModalOpen, setIsSettingsModalOpen] = useState<boolean>(false);
  const [isSettingsModalMounted, setIsSettingsModalMounted] = useState<boolean>(false);

  const connectionsPath = ROUTES.connections.path;
  const credentialsPath = ROUTES.credentials.path;
  const issuerStatePath = ROUTES.issuerState.path;
  const schemasPath = ROUTES.schemas.path;
  const identitiesPath = ROUTES.identities.path;
  const displayMethodsPath = ROUTES.displayMethods.path;
  const paymentOptionsPath = ROUTES.paymentOptions.path;
  const paymentRequestsPath = ROUTES.paymentRequests.path;

  const getSelectedKey = (): string[] => {
    if (
      matchRoutes(
        [
          { path: schemasPath },
          { path: ROUTES.importSchema.path },
          { path: ROUTES.schemaDetails.path },
        ],
        pathname
      )
    ) {
      return [schemasPath];
    } else if (
      matchRoutes(
        [
          { path: credentialsPath },
          { path: ROUTES.credentialDetails.path },
          { path: ROUTES.issueCredential.path },
          { path: ROUTES.linkDetails.path },
        ],
        pathname
      )
    ) {
      return [credentialsPath];
    } else if (
      matchRoutes([{ path: connectionsPath }, { path: ROUTES.connectionDetails.path }], pathname)
    ) {
      return [connectionsPath];
    } else if (matchRoutes([{ path: issuerStatePath }], pathname)) {
      return [issuerStatePath];
    } else if (
      matchRoutes(
        [
          { path: identitiesPath },
          { path: ROUTES.createIdentity.path },
          { path: ROUTES.identityDetails.path },
          { path: ROUTES.keyDetails.path },
          { path: ROUTES.createKey.path },
        ],
        pathname
      )
    ) {
      return [identitiesPath];
    } else if (
      matchRoutes(
        [
          { path: displayMethodsPath },
          { path: ROUTES.createDisplayMethod.path },
          { path: ROUTES.displayMethodDetails.path },
        ],
        pathname
      )
    ) {
      return [displayMethodsPath];
    } else if (
      matchRoutes(
        [
          { path: paymentOptionsPath },
          { path: ROUTES.createPaymentOption.path },
          { path: ROUTES.paymentOptionDetails.path },
        ],
        pathname
      )
    ) {
      return [paymentOptionsPath];
    } else if (
      matchRoutes(
        [{ path: paymentRequestsPath }, { path: ROUTES.paymentRequestDetils.path }],
        pathname
      )
    ) {
      return [paymentRequestsPath];
    }

    return [];
  };

  const onMenuClick = (path: string) => {
    onClick();
    navigate(path);
  };

  return (
    <>
      <Row
        className="menu-sider-layout"
        justify="space-between"
        style={{
          padding: isBreakpoint ? "32px 16px" : "96px 16px 32px",
        }}
      >
        <Col>
          <UserDisplay />

          <Menu
            items={[
              {
                icon: <IconSchema />,
                key: schemasPath,
                label: SCHEMAS,
                onClick: () => onMenuClick(schemasPath),
                title: "",
              },
              {
                icon: <IconDisplayMethod />,
                key: displayMethodsPath,
                label: DISPLAY_METHODS,
                onClick: () => onMenuClick(displayMethodsPath),
                title: "",
              },
              {
                icon: <IconCredentials />,
                key: credentialsPath,
                label: CREDENTIALS,
                onClick: () =>
                  onMenuClick(
                    generatePath(credentialsPath, {
                      tabID: CREDENTIALS_TABS[0].tabID,
                    })
                  ),
                title: "",
              },
              {
                icon: <IconConnections />,
                key: connectionsPath,
                label: CONNECTIONS,
                onClick: () => onMenuClick(connectionsPath),
                title: "",
              },
              {
                icon: <IconIssuerState />,
                key: issuerStatePath,
                label:
                  isAsyncTaskDataAvailable(status) && status.data ? (
                    <Space>
                      {ISSUER_STATE}
                      <Tag style={{ fontSize: 12 }}>Pending actions</Tag>
                    </Space>
                  ) : (
                    ISSUER_STATE
                  ),
                onClick: () => onMenuClick(issuerStatePath),
                title: "",
              },
              ...(paymentPagesEnabled
                ? [
                    {
                      icon: <IconPaymentOptions />,
                      key: paymentOptionsPath,
                      label: PAYMENT_OPTIONS,
                      onClick: () => onMenuClick(paymentOptionsPath),
                      title: "",
                    },
                    {
                      icon: <IconPaymentOptions />,
                      key: paymentRequestsPath,
                      label: PAYMENT_REQUESTS,
                      onClick: () => onMenuClick(paymentRequestsPath),
                      title: "",
                    },
                  ]
                : []),
            ]}
            selectedKeys={getSelectedKey()}
            style={{ marginTop: 16 }}
          />
        </Col>

        <Space direction="vertical" size={40}>
          <Menu
            items={[
              {
                icon: <IconIdentities />,
                key: identitiesPath,
                label: IDENTITIES,
                onClick: () => onMenuClick(identitiesPath),
                title: "",
              },
              {
                icon: <IconSettings />,
                key: "settings",
                label: (
                  <Typography.Link
                    onClick={() => {
                      setIsSettingsModalOpen(true);
                      setIsSettingsModalMounted(true);
                    }}
                  >
                    <Row justify="space-between">Settings</Row>
                  </Typography.Link>
                ),
              },
              {
                icon: <IconFile />,
                key: "documentation",
                label: (
                  <Typography.Link href={DOCS_URL} target="_blank">
                    <Row justify="space-between">
                      <span>Documentation</span>

                      <IconLink className="icon-secondary" height={16} />
                    </Row>
                  </Typography.Link>
                ),
              },
            ]}
            selectable={false}
            selectedKeys={getSelectedKey()}
          />
          {isBreakpoint && (
            <Space>
              <LogoLink />

              {buildTag && <Tag>{buildTag}</Tag>}
            </Space>
          )}
        </Space>
      </Row>

      {isSettingsModalMounted && (
        <SettingsModal
          afterOpenChange={setIsSettingsModalMounted}
          isOpen={isSettingsModalOpen}
          onClose={() => {
            setIsSettingsModalOpen(false);
          }}
          onSave={() => {
            setIsSettingsModalOpen(false);
            void message.success("IPFS gateway successfully changed");
          }}
        />
      )}
    </>
  );
}
