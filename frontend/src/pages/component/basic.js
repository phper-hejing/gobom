import { PureComponent, Fragment } from 'react';
import { REQUEST_HTTP, REQUEST_TCP } from '../../constant';
import { Radio, Input, Form, Popconfirm, InputNumber, Select } from 'antd';
import { PlusCircleOutlined, MinusCircleOutlined } from '@ant-design/icons';

const { Option } = Select;
const defaultMethods = ['GET', 'POST', 'PUT', 'DELETE'];
const formLayout = {
  labelCol: { span: 2 },
  wrapperCol: { span: 22 },
};

export default class basic extends PureComponent {
  static defaultProps = {
    requestType: REQUEST_HTTP,
  };

  constructor(props) {
    super(props);
    this.state = {
      conCurrent: 100,
      duration: 1,
      method: 'GET',

      addConcurrent:0,
      minusConcurrent:0,
      durationType:60,
      globalDisabled:false
    };
  }

  methodChange = e => {
    this.setState({
      method: e.target.value,
    });
  };

  concurrentChange = e => {
    this.setState({
      conCurrent: e.target.value
    })
  }

  durationChange = e => {
    this.setState({
      duration: e.target.value
    })
  }

  durationTypeChange = e => {
    this.setState({
      durationType: e.value
    })
  }

  minusConcurrentChange = e => {
    if (this.refs.minusConcurrent.input.value == undefined
      || this.refs.minusConcurrent.input.value == 0
      || this.refs.minusConcurrent.input.value >= this.state.conCurrent
    ) {
      return
    }
    this.setState( state => ({
      conCurrent: parseInt(this.state.conCurrent) - parseInt(this.refs.minusConcurrent.input.value)
    }))
  }

  addConcurrentChange = e => {
    if (this.refs.addConcurrent.input.value == undefined || this.refs.addConcurrent.input.value == 0) {
      return
    }
    this.setState( state => ({
      conCurrent: parseInt(state.conCurrent) + parseInt(this.refs.addConcurrent.input.value)
    }))
  }


  render() {
    const minusConcurrent = <div>Minus：<InputNumber disabled={this.state.globalDisabled} ref="minusConcurrent" style={{width:"100px"}} type="text" /></div>
    const addConcurrent = <div>Add：<InputNumber disabled={this.state.globalDisabled} ref="addConcurrent" style={{width:"100px"}} type="text" /></div>
    const concurrentPrefixDom = <div style={{cursor:"pointer"}}>
      <Popconfirm onConfirm={this.minusConcurrentChange} icon={null} title={minusConcurrent}>
        <MinusCircleOutlined />
      </Popconfirm>
    </div>;
    const concurrentSuffixDom = <div style={{cursor:"pointer"}}>
      <Popconfirm onConfirm={this.addConcurrentChange} icon={null} title={addConcurrent}>
        <PlusCircleOutlined />
      </Popconfirm>
    </div>;
    const durationSuffixDom = <div>
      <Select disabled={this.state.globalDisabled} defaultValue={this.state.durationType} onChange={this.durationTypeChange} style={{ width: 120 }}>
        <Option value={1}>秒</Option>
        <Option value={60}>分钟</Option>
        <Option value={60*60}>小时</Option>
      </Select>
    </div>

    return (
      <Fragment>
        <Form.Item label="Method" {...formLayout} labelAlign="left">
          <Radio.Group disabled={this.props.requestType != REQUEST_HTTP || this.state.globalDisabled} onChange={this.methodChange} value={this.props.requestType != REQUEST_HTTP ? "" : this.state.method}>
            {defaultMethods.map(function(item, index) {
              return (
                <Radio key={index} value={item}>
                  {item}
                </Radio>
              );
            })}
          </Radio.Group>
        </Form.Item>

        <Form.Item label="并发数" {...formLayout} labelAlign="left">
            <Input disabled={this.state.globalDisabled} prefix={concurrentPrefixDom} suffix={concurrentSuffixDom} onChange={this.concurrentChange} value={this.state.conCurrent} />
        </Form.Item>

        <Form.Item label="持续时间" {...formLayout} labelAlign="left">
          <Input disabled={this.state.globalDisabled} suffix={durationSuffixDom} onChange={this.durationChange} value={this.state.duration} />
        </Form.Item>
      </Fragment>
    );
  }
}
