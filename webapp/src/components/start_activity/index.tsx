import {connect} from 'react-redux';
import {bindActionCreators} from 'redux';
import {closeStartActivityModal} from '../../actions';
import {meetingTemplates, teams, isStartActivityModalVisible} from '../../selectors';

import StartActivity from './start_activity';

const mapStateToProps = (state, ownProps) => {
    return {
        visible: isStartActivityModalVisible(state),
        meetingTemplates: meetingTemplates(state),
    };
}

const mapDispatchToProps = (dispatch) => bindActionCreators({
    close: closeStartActivityModal
}, dispatch);

export default connect(mapStateToProps, mapDispatchToProps)(StartActivity);
