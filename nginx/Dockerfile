FROM nginx:latest

# Remove the default nginx.conf file
RUN rm /etc/nginx/conf.d/default.conf

# Copy the custom nginx.conf file to the container
COPY nginx.conf /etc/nginx/conf.d/

# Expose port 80
EXPOSE 80

# Start NGINX
CMD ["nginx", "-g", "daemon off;"]