import { PureComponent, Fragment } from 'react';
import { Table, Input, Select, InputNumber } from 'antd';

const { Option } = Select;
const defaultLen = 10;
const defaultType = 'int';

export default class body extends PureComponent {
  constructor(props) {
    super(props);

    let data =
      this.props.data && Object.keys(this.props.data).length != 0
        ? this.defaultData(this.props.data)
        : [
            {
              key: 1,
              name: '',
              type: defaultType,
              len: defaultLen,
              default: '',
              dynamic: '',
            },
          ];
    let defaultSelectedKeys = this.defaultSelectedKeys(data);
    let initKey = data.length;

    this.state = {
      initKey: initKey,
      columns: [
        {
          title: 'KEY',
          dataIndex: 'name',
          render: (text, record, index) => {
            return (
              <Input
                onChange={this.addOneRowData.bind(this, index, 'name')}
                size="small"
                value={text}
              />
            );
          },
        },
        {
          title: '类型',
          dataIndex: 'type',
          render: (text, record, index) => (
            <Select
              size="small"
              value={text}
              onChange={this.addOneRowData.bind(this, index, 'type')}
              style={{ width: 120 }}
            >
              <Option value="int">int</Option>
              <Option value="string">string</Option>
              <Option value="file">file</Option>
              <Option value="sendData">sendData</Option>
              <Option value="response">response</Option>
            </Select>
          ),
        },
        {
          title: '长度/范围',
          dataIndex: 'len',
          render: (text, record, index) => (
            <InputNumber
              value={text}
              size="small"
              onChange={this.addOneRowData.bind(this, index, 'len')}
              style={{ width: 100 }}
              type="text"
            />
          ),
        },
        {
          title: '默认值',
          dataIndex: 'default',
          render: (text, record, index) => (
            <Input
              onChange={this.addOneRowData.bind(this, index, 'default')}
              size="small"
              value={text}
            />
          ),
        },
        {
          title: '其他',
          dataIndex: 'dynamic',
          render: (text, record, index) => (
            <Input
              onChange={this.addOneRowData.bind(this, index, 'dynamic')}
              size="small"
              value={text}
            />
          ),
        },
      ],
      data: data,
      selectedRowKeys: defaultSelectedKeys,
    };
  }

  defaultData = propsData => {
    let data = this.mapTobody(propsData);
    data.push({
      key: data.length + 1,
      name: '',
      type: defaultType,
      len: defaultLen,
      default: '',
      dynamic: '',
    });
    return data;
  };

  defaultSelectedKeys = data => {
    let keys = [];
    for (let i = 1; i < data.length; i++) {
      keys.push(i);
    }
    return keys;
  };

  bodyToMap = (data, selectedRowKeys) => {
    let arr = [];
    Object.values(data).map(item => {
      if (
        item.name != '' &&
        (item.type != '' || item.dynamic != '') &&
        selectedRowKeys.includes(item.key)
      ) {
        arr.push({
          name: item.name,
          type: item.type,
          len: item.len == 0 ? 10 : item.len,
          default: item.default,
          dynamic: item.dynamic,
        });
      }
    });
    return arr;
  };

  mapTobody = data => {
    let arr = [];
    let i = 1;
    data.map(function(item) {
      arr.push({
        key: i,
        name: item.name,
        type: item.type,
        len: item.len,
        default: item.default,
        dynamic: item.dynamic,
      });
      i++;
    });
    return arr;
  };

  onSelectChange = selectedRowKeys => {
    if (this.props.render) {
      this.props.render(this.bodyToMap(this.state.data, selectedRowKeys));
    }
    this.setState({
      selectedRowKeys: selectedRowKeys,
    });
  };

  addOneRowData = (index, record, e) => {
    if (e == null) {
      return;
    }
    let data = this.state.data;
    let selectedRowKeys = [];
    let value = e.target ? e.target.value : e;
    data[index][record] = value;
    this.setState({
      data: data,
    });

    index += 1; // index从0开始 data的key从1开始 所以index+1

    if (index == this.state.initKey) {
      this.setState(state => ({
        initKey: ++state.initKey,
        data: data.concat([
          {
            key: state.initKey,
            name: '',
            type: defaultType,
            len: defaultLen,
            default: '',
            dynamic: '',
          },
        ]),
      }));
    }

    if (value == '' && record == 'name') {
      selectedRowKeys = this.state.selectedRowKeys.filter(key => {
        if (key != index) {
          return key;
        }
      });
      this.setState(state => {
        return {
          selectedRowKeys: selectedRowKeys,
        };
      });
    } else {
      selectedRowKeys = this.state.selectedRowKeys.concat([index]);
      this.setState(state => {
        return {
          selectedRowKeys: selectedRowKeys,
        };
      });
    }
    if (this.props.render) {
      this.props.render(this.bodyToMap(data, selectedRowKeys));
    }
  };

  componentDidUpdate(prevProps, prevState) {
    if (JSON.stringify(prevProps.data) != JSON.stringify(this.props.data)) {
      let data = this.defaultData(this.props.data);
      this.setState({
        data: data,
        selectedRowKeys: this.defaultSelectedKeys(data),
      });
    }
  }

  render() {
    const rowSelection = {
      onChange: this.onSelectChange,
      selectedRowKeys: this.state.selectedRowKeys,
    };
    return (
      <Fragment>
        <Table
          pagination={false}
          size="small"
          rowSelection={rowSelection}
          columns={this.state.columns}
          dataSource={this.state.data}
        />
      </Fragment>
    );
  }
}
