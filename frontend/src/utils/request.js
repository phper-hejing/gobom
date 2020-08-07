function checkStatus(response) {
    if (response.status >= 200 && response.status < 300) {
        return response;
    }

    const error = new Error(response.statusText);
    error.response = response;
    throw error;
}

export default async function request(url, options) {
    const response = await fetch(url, Object.assign({
      method: 'POST',
      headers: new Headers({
        'Content-Type': 'application/json',
      }),
    }, options));
    checkStatus(response);
    return response.json()
}
