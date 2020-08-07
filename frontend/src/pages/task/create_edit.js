import React, { PureComponent, Fragment } from 'react';
import { Form, Input, Button, Select, message, InputNumber } from 'antd';
import { connect } from 'dva';

const { Option } = Select;

const formItemLayout = {
  labelCol: { span: 4 },
  wrapperCol: { span: 20 },
};

const namespace = 'script';
const scriptModelState = state => {
  return { script: state[namespace] };
};
const scriptModelEvent = dispatch => {
  return {
    getScriptList: data => {
      dispatch({
        type: `${namespace}/getScriptList`,
      }).then(resp => {
        if (resp.msg != '') {
          message.error(resp.msg);
        }
      });
    },

    // ===============  task ===================
    add: data => {
      dispatch({
        type: `task/addTask`,
        payload: data,
      }).then(resp => {
        if (resp.msg != '') {
          message.error(resp.msg);
        } else {
          location.reload();
        }
      });
    },
    edit: data => {
      dispatch({
        type: `task/editTask`,
        payload: data,
      }).then(resp => {
        if (resp.msg != '') {
          message.error(resp.msg);
        } else {
          location.reload();
        }
      });
    },
  };
};
@connect(scriptModelState, scriptModelEvent)
export default class create_edit extends PureComponent {
  constructor(props) {
    super(props);
    this.state = {
      taskId: this.props.data ? this.props.data.taskId : '',
      name: this.props.data ? this.props.data.name : '',
      conCurrent: this.props.data ? this.props.data.conCurrent : 0,
      duration: this.props.data ? this.props.data.duration : 0,
      scriptId: this.props.data ? this.props.data.scriptId : false,
      isCreate: this.props.data && this.props.data.taskId ? false : true,
    };
    this.form = React.createRef();
    this.props.getScriptList();
  }

  nameChange = e => {
    this.setState({
      name: e.target.value,
    });
  };

  conCurrentChange = e => {
    this.setState({
      conCurrent: e,
    });
  };

  durationChange = e => {
    this.setState({
      duration: e,
    });
  };

  scriptIdChange = e => {
    this.setState({
      scriptId: e,
    });
  };

  handelSubmit = e => {
    console.log(e);
    if (this.state.isCreate) {
      this.props.add(this.state);
    } else {
      this.props.edit(this.state);
    }
  };

  render() {
    return (
      <Form {...formItemLayout} onFinish={this.handelSubmit}>
        <Form.Item
          initialValue={this.state.name}
          label="任务名称"
          name="name"
          rules={[{ required: true }]}
        >
          <Input onChange={this.nameChange} />
        </Form.Item>

        <Form.Item
          initialValue={this.state.conCurrent}
          label="并发数"
          name="conCurrent"
          rules={[{ required: true }]}
        >
          <InputNumber onChange={this.conCurrentChange} />
        </Form.Item>

        <Form.Item
          initialValue={this.state.duration}
          label="压测时间（s）"
          name="duration"
          rules={[{ required: true }]}
        >
          <InputNumber
            onChange={this.duration}
            onChange={this.durationChange}
          />
        </Form.Item>

        <Form.Item
          initialValue={this.state.scriptId}
          label="压测脚本"
          name="scriptId"
          rules={[{ required: true }]}
        >
          <Select onChange={this.scriptIdChange}>
            {this.props.script.scriptList
              ? this.props.script.scriptList.map(item => {
                  return (
                    <Option key={item.ID} value={item.ID}>
                      {item.name}
                    </Option>
                  );
                })
              : null}
          </Select>
        </Form.Item>

        <Form.Item wrapperCol={{ sm: { offset: 4 } }}>
          <Button type="primary" htmlType="submit">
            提交
          </Button>
        </Form.Item>
      </Form>
    );
  }
}
