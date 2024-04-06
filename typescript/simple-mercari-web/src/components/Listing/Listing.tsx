import React, { useState } from 'react';

const server = process.env.REACT_APP_API_URL || 'http://127.0.0.1:9000';

interface Prop {
  onListingCompleted?: () => void;
}

type formDataType = {
  name: string,
  category: string,
  image: string | File,
}

export const Listing: React.FC<Prop> = ({ onListingCompleted }) => {
  const initialState = {
    name: "",
    category: "",
    image: "",
  };
  const [values, setValues] = useState<formDataType>(initialState);
  const [imagePreview, setImagePreview] = useState<string | null>(null);

  const onValueChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setValues({
      ...values, [event.target.name]: event.target.value,
    })
  };
  const onFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files![0];
    setValues({
      ...values, [event.target.name]: file,
    });
    const reader = new FileReader();
    reader.onloadend = () => {
      setImagePreview(reader.result as string);
    };
    if (file) {
      reader.readAsDataURL(file);
    }
  };

  const onSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const data = new FormData();
    data.append('name', values.name);
    data.append('category', values.category);
    if (typeof values.image === 'object') {
      data.append('image', values.image);
    }

    fetch(server.concat('/items'), {
      method: 'POST',
      mode: 'cors',
      body: data,
    })
      .then(response => {
        console.log('POST status:', response.statusText);
        onListingCompleted && onListingCompleted();
      })
      .catch((error) => {
        console.error('POST error:', error);
      })
  };
  return (
    <div className='Listing'>
      <form onSubmit={onSubmit}>
        <div>
          <input type='text' name='name' id='name' placeholder='name' onChange={onValueChange} required />
          <input type='text' name='category' id='category' placeholder='category' onChange={onValueChange} />
          <input type='file' name='image' id='image' onChange={onFileChange} required />
          {imagePreview && <img src={imagePreview} alt="Preview" style={{width: '100px', height: '100px'}} />}
          <button type='submit'>List this item</button>
        </div>
      </form>
    </div>
  );
}
