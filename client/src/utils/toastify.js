import {toast} from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import * as React from 'react';

function Toast(type, hide_progress_bar, message) {
    const toastSwitch = (type, hide_progress_bar, message) => {
        const toastCfg = {
            position: "top-right",
            autoClose: 3000,
            hideProgressBar: hide_progress_bar,
            closeOnClick: true,
            pauseOnHover: true,
            draggable: true,
            progress: undefined,
        }
        switch(type) {
            case 'info':
                return (
                    toast.info(message, toastCfg)
                );
            case 'success':
                return (
                    toast.success(message, toastCfg)
                );
            case 'warn':
                return (
                    toast.warn(message, toastCfg)
                );
            case 'error':
                return (
                    toast.error(message, toastCfg)
                );
            default:
                return null;
        }
    }
    return (
        <div>
            {
                toastSwitch(type, hide_progress_bar, message)
            }
        </div>
    );
}

export default Toast