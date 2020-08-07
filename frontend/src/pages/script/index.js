import React, { Component, PureComponent } from 'react';
import { Table, Button, Popconfirm, message } from 'antd';
import { Modal } from 'antd';
import { connect } from 'dva';
import CraeteEdit from './create_edit';

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
    delScript: id => {
      dispatch({
        type: `${namespace}/delScript`,
        payload:id,
      }).then(resp => {
        if (resp.msg != '') {
          message.error(resp.msg);
        } else {
          location.reload()
        }
      });
    }
  };
};
@connect(scriptModelState, scriptModelEvent)
class index extends PureComponent {
  constructor(props) {
    super(props);
    this.state = {
      columns: [
        {
          title: '测试脚本名称',
          dataIndex: 'name',
        },
        {
          title: '协议',
          dataIndex: 'protocol',
        },
        {
          title: '操作',
          dataIndex: 'operation',
          render: (text, record) => {
            return (
              <div>
                <Button
                  onClick={this.edit.bind(this, record)}
                  type="primary"
                  size="small"
                  style={{ marginRight: 5 }}
                >
                  编辑
                </Button>
                <Popconfirm
                  title="Sure to delete?"
                  onConfirm={() => this.handleDelete(record)}
                >
                  <Button type="danger" size="small">
                    删除
                  </Button>
                </Popconfirm>
              </div>
            );
          },
        },
      ],
      data: [],
      editData: [],
      visible: false,
    };
  }

  handleDelete = record => {
    this.props.delScript(record.ID)
  };

  showModal = () => {
    this.setState({
      visible: !this.state.visible,
    });
  };

  updateCallback = () => {
    location.reload()
  }

  edit = record => {
    this.setState({
      editData: record,
    });
    this.showModal();
  };

  reload = () => {
    window.location.reload();
  };

  componentDidMount() {
    this.props.getScriptList();
  }

  componentDidUpdate(prevProps, prevState, snapshot) {
    // state和props.script.scriptList共用一个数组对象
    // 所以这里可以直接进行对比
    if (this.props.script.scriptList !== this.state.data) {
      this.setState({
        data: this.props.script.scriptList
      })
    }
  }

  title = () => {
    return (
      <div>
        <h1 style={{ fontWeight: 'bold' }}>测试脚本列表</h1>
        <Button
          onClick={() => {
            this.props.history.push('/script/create_edit?type=common');
          }}
          style={{ position: 'absolute', right: '150px', top: '15px' }}
          type="dashed"
        >
          创建普通脚本
        </Button>
        <Button
          onClick={() => {
            this.props.history.push('/script/create_edit?type=transaction');
          }}
          style={{ position: 'absolute', right: '10px', top: '15px' }}
          type="dashed"
        >
          创建事务脚本
        </Button>
      </div>
    );
  };

  render() {
    return (
      <div>
        <Modal
          width={1000}
          title="Modal"
          visible={this.state.visible}
          footer={null}
          title="编辑"
          onCancel={this.showModal}
        >
          <CraeteEdit updateCallback={this.updateCallback} script={this.state.editData}></CraeteEdit>
        </Modal>
        <Table
          rowKey="name" // key值从哪个字段取值
          locale={{ emptyText: '点击右上角按钮创建测试脚本' }}
          title={this.title}
          columns={this.state.columns}
          dataSource={this.state.data}
        />
      </div>
    );
  }
}
export default index;
