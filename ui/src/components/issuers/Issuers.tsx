import { Divider, Modal, Space, message } from "antd";
import { useCallback, useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";

import { createIssuer } from "src/adapters/api/issuers";
import { IssuerFormData } from "src/adapters/parsers/view";
import IconClose from "src/assets/icons/x.svg?react";
import { IssuerForm } from "src/components/issuers/IssuerForm";
import { IssuersTable } from "src/components/issuers/IssuersTable";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { useIssuerContext } from "src/contexts/Issuer";
import { ROUTES } from "src/routes";
import { isAsyncTaskDataAvailable } from "src/utils/async";
import { makeRequestAbortable } from "src/utils/browser";
import { ISSUERS, ISSUER_ADD } from "src/utils/constants";

export function Issuers() {
  const env = useEnvContext();
  const { fetchIssuers, issuersList } = useIssuerContext();

  const [messageAPI, messageContext] = message.useMessage();
  const [isModalOpen, setIsModalOpen] = useState(false);
  const navigate = useNavigate();

  const closeModal = () => {
    setIsModalOpen(false);
  };

  const handleAddIssuer = useCallback(() => {
    if (isAsyncTaskDataAvailable(issuersList)) {
      if (issuersList.data.length) {
        navigate(ROUTES.createIssuer.path);
      } else {
        setIsModalOpen(true);
      }
    }
  }, [issuersList, navigate]);

  const fetchData = useCallback(() => {
    const { aborter } = makeRequestAbortable(fetchIssuers);
    return aborter;
  }, [fetchIssuers]);

  const handleSubmit = useCallback(
    (formValues: IssuerFormData) =>
      void createIssuer({ env, payload: formValues }).then((response) => {
        if (response.success) {
          closeModal();
          fetchData();
          void messageAPI.success("Issuer added");
        } else {
          void messageAPI.error(response.error.message);
        }
      }),
    [fetchData, messageAPI, env]
  );

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  return (
    <>
      {messageContext}

      <SiderLayoutContent description="Description." title={ISSUERS}>
        <Divider />
        <Space direction="vertical" size="large">
          <IssuersTable handleAddIssuer={handleAddIssuer} />
          {isModalOpen && (
            <Modal
              centered
              closable
              closeIcon={<IconClose />}
              footer={null}
              maskClosable
              onCancel={closeModal}
              open
              title={ISSUER_ADD}
            >
              <IssuerForm onBack={closeModal} onSubmit={handleSubmit} />
            </Modal>
          )}
        </Space>
      </SiderLayoutContent>
    </>
  );
}