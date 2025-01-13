import { App, Card } from "antd";
import { useNavigate } from "react-router-dom";

import { createPaymentOption } from "src/adapters/api/payments";
import { notifyParseError } from "src/adapters/parsers";
import { PaymentOptionFormData, paymentOptionFormParser } from "src/adapters/parsers/view";
import { PaymentOptionForm } from "src/components/payments/PaymentOptionForm";
import { SiderLayoutContent } from "src/components/shared/SiderLayoutContent";
import { useEnvContext } from "src/contexts/Env";
import { useIdentityContext } from "src/contexts/Identity";
import { ROUTES } from "src/routes";

import { PAYMENT_OPTIONS_ADD_NEW } from "src/utils/constants";

export function CreatePaymentOption() {
  const env = useEnvContext();
  const { identifier } = useIdentityContext();

  const navigate = useNavigate();
  const { message } = App.useApp();

  const handleSubmit = (formValues: PaymentOptionFormData) => {
    const parsedFormData = paymentOptionFormParser.safeParse(formValues);

    if (parsedFormData.success) {
      return void createPaymentOption({
        env,
        identifier,
        payload: parsedFormData.data,
      }).then((response) => {
        if (response.success) {
          void message.success("Payment option added successfully");
          navigate(ROUTES.paymentOptions.path);
        } else {
          void message.error(response.error.message);
        }
      });
    } else {
      void notifyParseError(parsedFormData.error);
    }
  };

  return (
    <SiderLayoutContent
      description="Create a new payment option"
      showBackButton
      showDivider
      title={PAYMENT_OPTIONS_ADD_NEW}
    >
      <Card className="centered" title="Payment option details">
        <PaymentOptionForm
          initialValies={{
            description: "",
            name: "",
            paymentOptions: [],
          }}
          onSubmit={handleSubmit}
        />
      </Card>
    </SiderLayoutContent>
  );
}
