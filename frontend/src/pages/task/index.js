import React, { PureComponent, Component } from 'react';
import {
  Row,
  Col,
  Card,
  Statistic,
  List,
  Tag,
  Button,
  Table,
  message,
  Modal,
} from 'antd';
import {
  CheckOutlined,
  CloseCircleOutlined,
  ClockCircleOutlined,
  PlusCircleOutlined,
} from '@ant-design/icons';
import { connect } from 'dva';
import CraeteEdit from './create_edit';
import { WS_TASK_REPORT, WS_TASK_RUN, WS_TASK_STOP } from '../../constant';

const STATUS_NONE = 0;
const STATUS_WAIT = 1;
const STATUS_RUN = 2;
const STATUS_OVER = 3;
const STATUS_STOP = 4;
const STATUS_ERROR = 5;

const tagStatusColor = {
  [STATUS_NONE]: '',
  [STATUS_WAIT]: '#808080',
  [STATUS_RUN]: '#0eb83a',
  [STATUS_OVER]: '#f2be45',
  [STATUS_STOP]: '#4b5cc4',
  [STATUS_ERROR]: '#f20c00',
};
const tagStatusMsg = {
  [STATUS_NONE]: '--',
  [STATUS_WAIT]: '准备',
  [STATUS_RUN]: '运行',
  [STATUS_OVER]: '结束',
  [STATUS_STOP]: '停止',
  [STATUS_ERROR]: '错误',
};

const namespace = 'task';
const taskModelState = state => {
  return { task: state[namespace] };
};
const taskModelEvent = dispatch => {
  return {
    getTaskList: data => {
      dispatch({
        type: `${namespace}/getTaskList`,
        payload: data,
      }).then(resp => {
        if (resp.msg != '') {
          message.error(resp.msg);
        }
      });
    },
    delTask: function(record) {
      dispatch({
        type: `${namespace}/delTask`,
        payload: record.taskId,
      }).then(resp => {
        if (resp.msg != '') {
          message.error(resp.msg);
        } else {
          this.props.moveTask(record.key);
          this.setState({
            curTaskId: 0,
          });
        }
      });
    },
    moveTask: index => {
      dispatch({
        type: `${namespace}/moveTask`,
        payload: index,
      });
    },
    saveTask: data => {
      dispatch({
        type: `${namespace}/saveTask`,
        payload: data,
      });
    },
  };
};
const websocketModelState = state => {
  return { websocket: state['websocket'] };
};
const websocketModelEvent = dispatch => {
  return {
    registerHandel: handel => {
      dispatch({
        type: 'websocket/registerHandel',
        payload: handel,
      });
    },
    runTask: function(taskId) {
      if (this.props.websocket.ws && this.props.websocket.ws.readyState == 1) {
        if (this.state.startTasks[taskId]) {
          message.error('任务正在运行');
          return;
        }
        this.props.websocket.ws.send(
          JSON.stringify({
            type: WS_TASK_RUN,
            data: {
              taskId: taskId,
            },
          }),
        );
        this.setCurrentTask(taskId);
      } else {
        message.error('websocket连接失败，请刷新页面重试');
      }
    },
    stopTask: function(taskId) {
      if (this.props.websocket.ws && this.props.websocket.ws.readyState == 1) {
        this.props.websocket.ws.send(
          JSON.stringify({
            type: WS_TASK_STOP,
            data: {
              taskId: taskId,
            },
          }),
        );
        this.setCurrentTask(taskId);
      } else {
        message.error('websocket连接失败，请刷新页面重试');
      }
    },
    getReport: function(taskId) {
      if (this.props.websocket.ws && this.props.websocket.ws.readyState == 1) {
        this.props.websocket.ws.send(
          JSON.stringify({
            type: WS_TASK_REPORT,
            data: {
              taskId: taskId,
            },
          }),
        );
      } else {
        message.error('websocket连接失败，请刷新页面重试');
      }
    },
  };
};
@connect(taskModelState, taskModelEvent)
@connect(websocketModelState, websocketModelEvent)
export default class task extends PureComponent {
  constructor(props) {
    super(props);
    this.state = {
      modelShow: false,
      startTasks: {}, // 开启的任务列表 taskId:bool
      tasksStatus: {}, // 任务对应的状态 taskId:int
      curTaskId: 0, // 当前查看的任务id
      taskInterval: {}, // 任务对应的定时器,定时report
      editData: {},
      columns: [
        {
          title: '任务名称',
          dataIndex: 'name',
        },
        {
          title: '创建时间',
          dataIndex: 'createdAt',
        },
        {
          title: '状态',
          dataIndex: 'status',
          render: (status, record) => {
            return (
              <Tag
                color={
                  tagStatusColor[
                    this.state.tasksStatus[record.taskId]
                      ? this.state.tasksStatus[record.taskId]
                      : status
                  ]
                }
              >
                {
                  tagStatusMsg[
                    this.state.tasksStatus[record.taskId]
                      ? this.state.tasksStatus[record.taskId]
                      : status
                  ]
                }
              </Tag>
            );
          },
        },
        {
          title: '操作',
          dataIndex: '',
          render: record => {
            return (
              <div>
                <a
                  onClick={this.props.runTask.bind(this, record.taskId)}
                  disabled={this.state.startTasks[record.taskId] ? true : false}
                >
                  启动
                </a>
                &nbsp;
                <a
                  onClick={this.props.stopTask.bind(this, record.taskId)}
                  disabled={this.state.startTasks[record.taskId] ? false : true}
                >
                  停止
                </a>
                &nbsp;
                {/*<a onClick={this.setCurrentTask.bind(this, record.taskId)}>*/}
                {/*查看*/}
                {/*</a>*/}
                {/*&nbsp;*/}
                <a
                  onClick={this.editTask.bind(this, record)}
                  disabled={this.state.startTasks[record.taskId] ? true : false}
                >
                  编辑
                </a>
                &nbsp;
                <a
                  onClick={this.props.delTask.bind(this, record)}
                  disabled={this.state.startTasks[record.taskId] ? true : false}
                >
                  删除
                </a>
                &nbsp;
              </div>
            );
          },
        },
      ],
    };
  }

  componentDidMount() {
    this.props.getTaskList();
    this.props.registerHandel({
      key: WS_TASK_REPORT,
      callback: (ws, resp) => {
        if (resp.data.task.status != STATUS_RUN) {
          this.delTaskInterval(resp.data.task.taskId);
          this.delRunTasks(resp.data.task.taskId);
          this.setTasksStatus(resp.data.task.taskId, resp.data.task.status);
        }
        this.props.saveTask(resp.data);
      },
    });
    this.props.registerHandel({
      key: WS_TASK_RUN,
      callback: (ws, resp) => {
        if (resp.error != '') {
        }
        let taskId = resp.data.taskId;
        this.setRunTask(taskId);
        this.setTasksStatus(taskId, STATUS_RUN);
        this.intervalReport(taskId);
      },
    });
    this.props.registerHandel({
      key: WS_TASK_STOP,
      callback: (ws, resp) => {
        let taskId = resp.data.taskId;
        this.delRunTasks(taskId);
        if (resp.error != '') {
          this.setTasksStatus(taskId, STATUS_OVER);
        } else {
          this.setTasksStatus(taskId, STATUS_STOP);
        }
      },
    });
  }

  componentDidUpdate(prevProps, prevState) {
    if (this.props.task.taskList.length != prevProps.task.taskList.length) {
      let taskList = {};
      for (let index in this.props.task.taskList) {
        if (this.props.task.taskList[index].status == 2) {
          taskList[this.props.task.taskList[index].taskId] = true;
          setTimeout(() => {
            // 等待websocket初始化完成
            this.intervalReport(this.props.task.taskList[index].taskId);
          }, 1000);
        }
      }
      this.setRunTasks(taskList);
    }
  }

  intervalReport = taskId => {
    let interval = setInterval(() => {
      this.props.getReport.call(this, taskId);
    }, 1000);
    this.setTaskInterval(taskId, interval);
  };

  setCurrentTask = taskId => {
    for (let index in this.props.task.taskList) {
      if (this.props.task.taskList[index]['taskId'] == taskId) {
        this.setState({
          curTaskId: index,
        });
      }
    }
  };

  setRunTask = taskId => {
    if (!taskId) {
      console.debug('taskId is null');
      return;
    }
    let taskList = JSON.parse(JSON.stringify(this.state.startTasks));
    taskList[taskId] = true;
    this.setState({
      startTasks: taskList,
    });
  };

  setRunTasks = taskList => {
    if (!taskList || Object.keys(taskList).length == 0) {
      console.debug('taskIdList is null');
      return;
    }
    this.setState({
      startTasks: taskList,
    });
  };

  delRunTasks = taskId => {
    let startTasks = JSON.parse(JSON.stringify(this.state.startTasks));
    if (startTasks[taskId]) {
      startTasks[taskId] = false;
    }
    this.setState({
      startTasks: startTasks,
    });
  };

  setTasksStatus = (taskId, status) => {
    let tasksStatus = JSON.parse(JSON.stringify(this.state.tasksStatus));
    tasksStatus[taskId] = status;
    this.setState({
      tasksStatus: tasksStatus,
    });
  };

  isModelShow = () => {
    this.setState({
      modelShow: !this.state.modelShow,
    });
  };

  setTaskInterval = (taskId, interval) => {
    let taskInterval = JSON.parse(JSON.stringify(this.state.taskInterval));
    if (!taskInterval[taskId]) {
      taskInterval[taskId] = interval;
    }
    this.setState({
      taskInterval: taskInterval,
    });
  };

  delTaskInterval = taskId => {
    let taskInterval = JSON.parse(JSON.stringify(this.state.taskInterval));
    let interval = taskInterval[taskId];
    if (interval) {
      clearInterval(interval);
      delete taskInterval[taskId];
    }
    this.setState({
      taskInterval: taskInterval,
    });
  };

  editTask = record => {
    this.isModelShow();
    this.setState({
      editData: record,
    });
  };

  render() {
    return (
      <div>
        <Row>
          <Col span={16}>
            <div>
              <Row>
                <Col span={6}>
                  <Card
                    title={false}
                    size="small"
                    bordered={false}
                    style={{ width: 300 }}
                  >
                    <div>
                      <Statistic
                        title="任务名称"
                        value={
                          this.props.task.taskList[this.state.curTaskId]
                            ? this.props.task.taskList[this.state.curTaskId]
                                .name
                            : '--'
                        }
                      />
                    </div>
                  </Card>
                </Col>
                <Col span={6}>
                  <Card
                    title={false}
                    size="small"
                    bordered={false}
                    style={{ width: 300 }}
                  >
                    <div>
                      <Statistic
                        title="并发数"
                        value={
                          this.props.task.taskList[this.state.curTaskId]
                            ? this.props.task.taskList[this.state.curTaskId]
                                .conCurrent
                            : '--'
                        }
                      />
                    </div>
                  </Card>
                </Col>
                <Col span={6}>
                  <Card
                    title={false}
                    size="small"
                    bordered={false}
                    style={{ width: 300 }}
                  >
                    <div>
                      <Statistic
                        title="剩余时间（ms）"
                        value={
                          this.props.task.taskList[this.state.curTaskId]
                            ? this.props.task.taskList[this.state.curTaskId]
                                .duration == 0
                              ? '手动停止'
                              : this.props.task.taskList[this.state.curTaskId]
                                  .duration
                            : '--'
                        }
                      />
                    </div>
                  </Card>
                </Col>
                <Col span={6}>
                  <Card
                    title={false}
                    size="small"
                    bordered={false}
                    style={{ width: 300 }}
                  >
                    <div>
                      <Statistic
                        title="任务状态"
                        value={
                          tagStatusMsg[
                            this.props.task.taskList[this.state.curTaskId]
                              ? this.props.task.taskList[this.state.curTaskId]
                                  .status
                              : STATUS_NONE
                          ]
                        }
                        valueStyle={{
                          color:
                            tagStatusColor[
                              this.props.task.taskList[this.state.curTaskId]
                                ? this.props.task.taskList[this.state.curTaskId]
                                    .status
                                : STATUS_NONE
                            ],
                        }}
                      />
                    </div>
                  </Card>
                </Col>

                <Col span={6}>
                  <Card
                    title={false}
                    size="small"
                    bordered={false}
                    style={{ width: 300 }}
                  >
                    <div>
                      <Statistic
                        title="成功数"
                        value={
                          this.props.task.taskList[this.state.curTaskId]
                            ? this.props.task.taskList[this.state.curTaskId]
                                .successNum
                            : '--'
                        }
                      />
                    </div>
                  </Card>
                </Col>
                <Col span={6}>
                  <Card
                    title={false}
                    size="small"
                    bordered={false}
                    style={{ width: 300 }}
                  >
                    <div>
                      <Statistic
                        title="失败数"
                        value={
                          this.props.task.taskList[this.state.curTaskId]
                            ? this.props.task.taskList[this.state.curTaskId]
                                .failureNum
                            : '--'
                        }
                      />
                    </div>
                  </Card>
                </Col>
                <Col span={6}>
                  <Card
                    title={false}
                    size="small"
                    bordered={false}
                    style={{ width: 300 }}
                  >
                    <div>
                      <Statistic
                        title="最长请求时间（ms）"
                        value={
                          this.props.task.taskList[this.state.curTaskId]
                            ? this.props.task.taskList[this.state.curTaskId]
                                .maxTime
                            : '--'
                        }
                      />
                    </div>
                  </Card>
                </Col>
                <Col span={6}>
                  <Card
                    title={false}
                    size="small"
                    bordered={false}
                    style={{ width: 300 }}
                  >
                    <div>
                      <Statistic
                        title="最短请求时间（ms）"
                        value={
                          this.props.task.taskList[this.state.curTaskId]
                            ? this.props.task.taskList[this.state.curTaskId]
                                .minTime
                            : '--'
                        }
                      />
                    </div>
                  </Card>
                </Col>
              </Row>
            </div>
          </Col>
          <Col span={8}>
            <Button
              onClick={this.isModelShow}
              icon={<PlusCircleOutlined />}
              size="small"
              type="circle"
              style={{ position: 'absolute', top: 8, right: 10, zIndex: 99 }}
            ></Button>
            <Table
              size="small"
              columns={this.state.columns}
              dataSource={this.props.task.taskList}
              onRow={record => {
                return {
                  onClick: event => {
                    this.setCurrentTask(record.taskId);
                  },
                };
              }}
            />
            <Modal
              width={1000}
              title="create_edit"
              onCancel={this.isModelShow}
              visible={this.state.modelShow}
              footer={null}
            >
              <CraeteEdit data={this.state.editData}></CraeteEdit>
            </Modal>
          </Col>
        </Row>
      </div>
    );
  }
}
