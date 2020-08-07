import { Component, Fragment } from 'react';
import { Table, Input } from 'antd';

export default class kv extends Component {
  constructor(props) {
    super(props);
    let data =
      this.props.data && Object.keys(this.props.data).length != 0
        ? this.defaultData(this.props.data)
        : [
            {
              key: 1,
              index: '',
              value: '',
            },
          ];
    let defaultSelectedKeys = this.defaultSelectedKeys(data);
    let initKey = data.length;

    this.state = {
      initKey: initKey,
      columns: [
        {
          title: 'KEY',
          dataIndex: 'index',
          render: (text, record, index) => (
            <Input
              onChange={this.addOneRowData.bind(this, index, 'index')}
              size="small"
              value={text}
            />
          ),
        },
        {
          title: 'VALUE',
          dataIndex: 'value',
          render: (text, record, index) => (
            <Input
              onChange={this.addOneRowData.bind(this, index, 'value')}
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

  defaultSelectedKeys = data => {
    let keys = [];
    for (let i = 1; i <= data.length; i++) {
      if (i == data.length) {
        continue;
      }
      keys.push(i);
    }
    return keys;
  };

  defaultData = propsData => {
    let data = this.mapTokv(propsData);
    data.push({
      key: data.length + 1,
      index: '',
      value: '',
    });
    return data;
  };

  kvToMap = (data, selectedRowKeys) => {
    let arr = {};
    Object.values(data).map(item => {
      if (item.index != '' && selectedRowKeys.includes(item.key)) {
        arr[item.index] = item.value;
      }
    });
    return arr;
  };

  mapTokv = data => {
    let arr = [];
    let i = 1;
    Object.keys(data).map(index => {
      arr.push({
        key: i,
        index: index,
        value: data[index],
      });
      i++;
    });
    return arr;
  };

  onSelectChange = selectedRowKeys => {
    this.props.render(this.kvToMap(this.state.data, selectedRowKeys));
    this.setState({
      selectedRowKeys: selectedRowKeys,
    });
  };

  addOneRowData = (index, record, e) => {
    let data = this.state.data;
    let selectedRowKeys = [];
    data[index][record] = e.target.value;
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
            index: '',
            value: '',
          },
        ]),
      }));
    }

    if (e.target.value == '') {
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
      this.props.render(this.kvToMap(data, selectedRowKeys));
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
