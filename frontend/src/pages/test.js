import { Component, PureComponent } from 'react';

export  default class test extends PureComponent {

  constructor(props) {
    super(props);
    this.state = {
     a: {
       b: {
         c: 123
       }
     }
    }
  }

  handelClick = collapsed => {
    let data = Object.assign({}, this.state)
    data.a.b.c = 321
    this.setState(data)
    this.forceUpdate()
  }


  render() {
    return <div>
      <button onClick={this.handelClick.bind(this, this.state.collapsed)}>测试</button>
      {this.state.a.b.c}
    </div>
  }
}
