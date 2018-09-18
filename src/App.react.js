import React, { Component } from 'react';
import { TextField, Button } from '@material-ui/core';

import './App.css';

import Home from './components/Home.react';
import w2mService from './services/w2mService';

export default class App extends Component {
  constructor(props) {
    super(props);
    this.state = {
      id: "",
      code: "",
      timeZone: "us-LA", // TODO lmao use moment pls

      availability: null,
    };
  }

  handleChange = field => event => {
    this.setState({
      [field]: event.target.value
    });
  }

  handleSetIdCode() {
    const { id, code } = this.state;
    w2mService.getAvailability(id, code)
    .then((availability) => {
      this.setState({availability});
    });
  }

  render() {
    let content;
    if (this.state.availability) {
      content = (
        <Home {...this.state} />
      );
    } else {
      content = (
        <div>
          <TextField label="ID" value={this.state.id} onChange={this.handleChange('id')}/>
          <TextField label="Code" value={this.state.code} onChange={this.handleChange('code')}/>
          <Button onClick={this.handleSetIdCode.bind(this)}>Set ID and Code</Button>
        </div>
      );
    }

    // TODO use Material UI
    // TODO have them copy URL and massage it
    return (
      <div className="container">
        {content}
      </div>
    );
  }
}
