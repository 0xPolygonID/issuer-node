import { DatePicker, Form, FormItemProps } from "antd";

import { DATE_VALIDITY_MESSAGE } from "src/utils/constants";

export function Datetime({
  formItemProps,
  showTime,
}: {
  formItemProps: FormItemProps;
  showTime: boolean;
}): JSX.Element {
  return (
    <Form.Item {...formItemProps} rules={[{ message: DATE_VALIDITY_MESSAGE }]}>
      <DatePicker showTime={showTime} />
    </Form.Item>
  );
}
