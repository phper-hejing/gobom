import request from '../utils/request';
import { REQUEST_URL } from '../constant';

const scriptDataUrl = REQUEST_URL + '/script';
const scriptAddUrl = REQUEST_URL + '/script/add';
const scriptEditUrl = REQUEST_URL + '/script/edit';
const scriptDelUrl = REQUEST_URL + '/script/delete';
const scriptTestUrl = REQUEST_URL + '/script/test';

export default {
  namespace: 'script',
  state: {
    script: {},
    scriptList: [],
  },
  effects: {
    *getScriptList(_, { call, put }) {
      const resp = yield call(request, scriptDataUrl);
      if (resp.msg == '') {
        yield put({ type: 'setScriptList', payload: resp.data });
      }
      return resp;
    },
    *getScript({ payload: id }, { call, put }) {
      const script = yield call(request, scriptDataUrl, {
        body: {
          id: id,
        },
      });
      yield put({ type: 'setScript', payload: script });
    },
    *addScript({ payload: script }, { call, put }) {
      script.data = JSON.stringify(script.data);
      return yield call(request, scriptAddUrl, {
        body: JSON.stringify(script),
      });
    },
    *editScript({ payload: script }, { call, put }) {
      script.data = JSON.stringify(script.data);
      return yield call(request, scriptEditUrl, {
        body: JSON.stringify(script),
      });
    },
    *delScript({ payload: id }, { call, put }) {
      return yield call(request, scriptDelUrl, {
        body: JSON.stringify({
          id: id,
        }),
      });
    },
    *testScript({ payload: script }, { call, put }) {
      script.data = JSON.stringify(script.data);
      return yield call(request, scriptTestUrl, {
        body: JSON.stringify(script),
      });
    },
  },
  reducers: {
    setScriptList(state, { payload: scriptList }) {
      return {
        ...state,
        scriptList: scriptList,
      };
    },
    setScript(state, { payload: script }) {
      return {
        ...state,
        script: script,
      };
    },
  },
};
