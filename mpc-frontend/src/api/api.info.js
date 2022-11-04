import API from './api';

export const getOnline = async () => {
  try {
    const response = await API.get(`/online`);
    return response.data;
  } catch (e) {
    console.error("Unable to load online participants");
    return null;
  }
};

export const getConfigs = async () => {
  try {
    const response = await API.get(`/configs`);
    if (response.status !== 200) {
      return {
        error: true,
        data: response.data
      }
    }
    return {
      error: false,
      data: response.data
    };
  } catch (e) {
    return {
      error: true,
      data: JSON.stringify(e.response.data)
    };
  }
};
