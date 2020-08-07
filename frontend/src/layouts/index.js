import React, { PureComponent } from 'react';
import { Layout, Menu, Spin } from 'antd';
import { history } from 'umi';
import 'antd/dist/antd.css';
import styles from './index.css';
import { BarsOutlined, BugOutlined, RedoOutlined } from '@ant-design/icons';
import ws from '../models/websocket';
import { REQUEST_WS_URL } from '../constant';
import { connect } from 'dva';

const { Header, Content, Footer, Sider } = Layout;

const namespace = 'websocket';
const websocketModelState = state => {
  return { websocket: state[namespace] };
};
const websocketModelEvent = dispatch => {
  return {
    init: function(url) {
      dispatch({
        type: `${namespace}/init`,
        payload: {
          url: url,
          callback: () => {
            this.setState({
              websocketLoading: false,
            });
          },
        },
      });
    },
  };
};
@connect(websocketModelState, websocketModelEvent)
export default class index extends PureComponent {
  constructor(props) {
    super(props);
    this.state = {
      collapsed: false,
      logoSize: 2,
      logoLocation: 'left',
      logoMargin: '24px',
      curPath: '',
      websocketLoading: true,

      ws: ws.ws,
    };
  }

  linkTo = item => {
    history.push(item.key);
  };

  onCollapse = collapsed => {
    var st = Object.assign({}, this.state);
    if (collapsed === true) {
      st.logoSize = 1.5;
      st.logoLocation = 'center';
      st.logoMargin = '0px';
    } else {
      st.logoSize = 2;
      st.logoLocation = 'left';
      st.logoMargin = '24px';
    }
    st.collapsed = collapsed;
    this.setState(st);
  };

  componentDidMount() {
    this.props.init.call(this, REQUEST_WS_URL + '/task/ws');
  }

  render() {
    let tip = (
      <div>
        <span>websocket加载中...</span>
      </div>
    );
    return (
      <Spin spinning={this.state.websocketLoading} tip={tip}>
        <Layout style={{ minHeight: '100vh' }}>
          <Sider
            collapsible
            collapsed={this.state.collapsed}
            onCollapse={this.onCollapse}
          >
            <div
              style={{ textAlign: this.state.logoLocation }}
              className={styles.logo}
            >
              <span
                style={{
                  fontSize: `${this.state.logoSize}rem`,
                  marginLeft: this.state.logoMargin,
                }}
                className={styles.title}
              >
                GoBom
              </span>
            </div>
            <Menu
              theme="dark"
              onClick={this.linkTo}
              defaultSelectedKeys={[this.state.curPath]}
              mode="inline"
            >
              <Menu.Item key="/task">
                <BarsOutlined />
                <span>测试任务</span>
              </Menu.Item>
              <Menu.Item key="/script">
                <BugOutlined />
                <span>测试脚本</span>
              </Menu.Item>
            </Menu>
          </Sider>
          <Layout>
            <Header style={{ background: '#fff', padding: 0 }} />

            <Content style={{ margin: '16px' }}>
              <div style={{ padding: 15, background: '#fff', minHeight: 360 }}>
                {this.props.children}
              </div>
            </Content>
            <Footer style={{ textAlign: 'center' }}>
              GO BOM ©2020 Created by Ant UED
            </Footer>
          </Layout>
        </Layout>
      </Spin>
    );
  }
}
