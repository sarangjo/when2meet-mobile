// @flow
import React from 'react';
import { TextField, Button } from '@material-ui/core';

import './services/w2mService';

export default class Home extends React.Component {
    constructor(props) {
        super(props);
        this.state = {
            name: "",
            password: "",
            userAvailability: [],

            userId: null,
        };
    }

    handleLogin() {
        const { id, name, password } = this.state;
        w2mService.login(id, name, password)
        .then(function(userId) {
            this.setState({userId})
            // TODO populate input availability based on the logged-in user, i.e. pull out this user's availability and plop it in
        });
    }

    // TODO should be automatic?
    handleSave() {
        const { userId, id, userAvailability };
        w2mService.saveTimes(userId, id, userAvailability)
        .then(function() {
            //- TODO: WHAT DOES THIS RETURN???
            // this.setState(??????????????);
            // maybe alert? lel
        });
    }

    render() {
        // TODO convert to UI
        const availabilityView = JSON.encode(this.props.availability);

        let content;
        if (!this.state.userId) {
            content = (
                <div className="login">
                    <TextField label="Name" value={this.state.name} onChange={this.handleChange('name')}/>
                    <TextField label="Password" value={this.state.password} onChange={this.handleChange('password')}/>
                    <Button onClick={this.handleLogin}>Login</Button>
                </div>
            );
        } else {
            content = (
                <div>
                    User's availability goes here
                    <Button onClick={this.handleSave}>Save</Button>
                </div>
            );
        }

        return (
            <div>
                {availabilityView}
                {content}
            </div>
        );
    }
}

// TODO set proptypes
// id, code, timeZone, availability
