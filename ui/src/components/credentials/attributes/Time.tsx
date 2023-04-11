import { Form, FormItemProps, TimePicker } from "antd";

import { TIME_VALIDITY_MESSAGE } from "src/utils/constants";

export function Time({ formItemProps }: { formItemProps: FormItemProps }): JSX.Element {
  return (
    <Form.Item {...formItemProps} rules={[{ message: TIME_VALIDITY_MESSAGE }]}>
      <TimePicker />
    </Form.Item>
  );
}
