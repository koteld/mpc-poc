import API from './api';

export const startDKG = async () => {
  try {
    const response = await API.post(`/keys/generate`);
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

export const startDKF = async (address) => {
  try {
    const response = await API.post(`/keys/refresh`, {
      address
    });
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

export const sendETH = async (address, to, amount) => {
  try {
    const response = await API.post(`/sendeth`, {
      address,
      to,
      amount
    });
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
