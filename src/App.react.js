import React, { Component } from 'react';
import { BrowserRouter, Switch, Route } from 'react-router-dom';
import { Button } from '@material-ui/core';

import './App.css';

import './services/w2mService';

export default class App extends Component {
    constructor(props) {
        super(props);
        this.state = {
            id: 0,
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
        // TODO use Material UI
        // TODO have them copy URL and massage it
        return (
            <div className="container">
                <div className="main">
                    <TextField label="ID" value={this.state.id} onChange={this.handleChange('id')}/>
                    <TextField label="Code" value={this.state.code} onChange={this.handleChange('code')}/>
                    <Button onClick={this.handleSetIdCode}>Set ID and Code</Button>
                    {availabilityView}
                </div>
            </div>
        )
    }
}
